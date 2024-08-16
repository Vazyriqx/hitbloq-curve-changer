package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"hitbloq-curve-changer/api"
	"hitbloq-curve-changer/hitBloqtypes"
	"math"
	"os"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var (
	_verbose bool
	_debug   bool
)

var rootCmd = &cobra.Command{
	Use:   "hitbloq-curve-changer <pool> <new-curve-json>",
	Short: "Update star ratings to preserve top CR with a new curve",
	Long: `update the star ratings for a pool to preserve the top CR for a new linear curve
Examples:
hitbloq-curve-changer poodles '{"type": "basic", "baseline": 0.78, "cutoff": 0.5, "exponential": 2.5}'
hitbloq-curve-changer poodles '{"type": "linear", "points": [[0.0, 0.0], [0.8, 0.5], [1.0, 1.0]]}'`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		poolName := args[0]
		newCurve, err := stringToCurve(args[1])
		if err != nil {
			return err
		}

		if err := validateCurve(newCurve); err != nil {
			return err
		}

		leaderboardIDs, err := fetchLeaderboardIDs(poolName)
		if err != nil {
			return err
		}

		newCommands, revertCommands, err := generateCommands(leaderboardIDs, poolName, newCurve)
		if err != nil {
			return err
		}

		if err := writeCommandsToFile("newCommands.txt", newCommands); err != nil {
			return err
		}
		if err := writeCommandsToFile("revertCommands.txt", revertCommands); err != nil {
			return err
		}

		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		if err.Error() != "accepts 2 arg(s), received 0" {
			log.Fatal(err)
		}
	}
}

func configureLoggingLevel() {
	log.SetFormatter(&log.TextFormatter{TimestampFormat: "2006-01-02 15:04:05", FullTimestamp: true})
	log.SetLevel(log.WarnLevel)
	if _verbose {
		log.SetLevel(log.InfoLevel)
	}
	if _debug {
		log.SetLevel(log.DebugLevel)
	}
}

func init() {
	cobra.OnInitialize(configureLoggingLevel)
	rootCmd.SilenceErrors = true
	rootCmd.PersistentFlags().BoolVarP(&_verbose, "verbose", "v", false, "enable verbose logging")
	rootCmd.PersistentFlags().BoolVar(&_debug, "debug", false, "enable debug logging")
}

func stringToCurve(input string) (*hitBloqtypes.CRCurve, error) {
	var result *hitBloqtypes.CRCurve
	err := json.Unmarshal([]byte(input), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func validateCurve(curve *hitBloqtypes.CRCurve) error {
	switch curve.Type {
	case "basic":
		return validateExponentialCurve(curve)
	case "linear":
		return validateLinearCurve(curve)
	default:
		return errors.New("invalid curve")
	}
}

func validateExponentialCurve(curve *hitBloqtypes.CRCurve) error {
	if curve.Cutoff != 0 {
		return nil
	} else {
		return errors.New("the cutoff may not be 0")
	}
}

func validateLinearCurve(curve *hitBloqtypes.CRCurve) error {
	if curve.Points[0][0] != 0.0 ||
		curve.Points[0][1] != 0.0 ||
		curve.Points[len(curve.Points)-1][0] != 1.0 ||
		curve.Points[len(curve.Points)-1][1] != 1.0 {
		return errors.New("the curve must start and end with [0.0,0.0], [1.0,1.0] respectively")
	}
	if len(curve.Points) > 15 {
		return errors.New("you may not have more than 15 points in your curve")
	}
	for i := 0; i < len(curve.Points)-1; i++ {
		if curve.Points[i][0] >= curve.Points[i+1][0] {
			return errors.New("the x values for every point must be unique and ascending in order")
		}
	}
	return nil
}

func applyCurve(firstPlaceAcc float64, curve *hitBloqtypes.CRCurve) (float64, error) {
	switch curve.Type {
	case "basic":
		return applyExponentialCurve(firstPlaceAcc, curve), nil
	case "linear":
		return applyLinearCurve(firstPlaceAcc, curve), nil
	default:
		return -1, errors.New("non valid curve type")
	}
}

func applyExponentialCurve(acc float64, curve *hitBloqtypes.CRCurve) float64 {
	acc = math.Min(100, acc)
	baseline := curve.Baseline
	cutoff := curve.Cutoff
	exponential := curve.Exponential
	if acc < baseline {
		return acc / 100 * cutoff
	} else {
		return acc/100*cutoff + (1-cutoff)*((acc-baseline)/math.Pow((100-baseline), exponential))
	}
}

func applyLinearCurve(acc float64, curve *hitBloqtypes.CRCurve) float64 {
	acc = math.Min(100, acc) / 100
	i := 0
	for range curve.Points {
		if acc < curve.Points[i][0] {
			break
		}
		i++
	}
	middle_dis := (acc - curve.Points[i-1][0]) / (curve.Points[i][0] - curve.Points[i-1][0])
	return curve.Points[i-1][1] + middle_dis*(curve.Points[i][1]-curve.Points[i-1][1])

}

func findNewStarRating(scoreWeight, desiredCR float64) float64 {
	starBonusMultiplier := 50.0
	return desiredCR / starBonusMultiplier / scoreWeight
}

func acc(score hitBloqtypes.Score, leaderboardInfo hitBloqtypes.LeaderboardInfo) float64 {
	var maxScore int
	if leaderboardInfo.Notes > 13 {
		maxScore = leaderboardInfo.Notes*920 - 7245
	} else {
		maxScores := []int{115, 345, 575, 805, 1035, 1495, 1955, 2415, 2875, 3335, 3795, 4255, 4715}
		maxScore = maxScores[leaderboardInfo.Notes-1]
	}
	return float64(score.Score) / float64(maxScore) * 100.0
}

func writeCommandsToFile(filename, content string) error {
	if err := os.WriteFile(filename, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write commands to file: %s: %w", filename, err)
	}
	return nil
}

func fetchLeaderboardIDs(poolName string) ([]string, error) {
	log.Info("Fetching leaderbnoard IDs\n")
	leaderboardIDs, err := api.GetAllLeaderboardIDs(poolName)
	if err != nil {
		return nil, err
	}
	return leaderboardIDs, nil
}

func generateCommands(leaderboardIDs []string, poolName string, curve *hitBloqtypes.CRCurve) (string, string, error) {
	tmpl := `{{ bar . "[" "#" "#" "-" "]"}} {{percent .}} | {{counters .}} | Elapsed: {{etime . | yellow}} | Remaining: {{rtime . | blue}}`
	log.Info("Generating reweight commands\n")
	loadingbar := pb.New(len(leaderboardIDs))
	loadingbar.SetTemplateString(tmpl)
	loadingbar.SetRefreshRate(time.Second / 2)
	loadingbar.Start()

	var revert strings.Builder
	var new strings.Builder
	for _, id := range leaderboardIDs {
		leaderboardInfo, err := api.GetLeaderboardInfo(id)
		if err != nil {
			return "", "", err
		}

		scores, err := api.GetScores(id, 0)
		if err != nil {
			return "", "", err
		}

		firstPlaceAcc := acc(scores[0], *leaderboardInfo)
		firstPlaceWeighting, err := applyCurve(firstPlaceAcc, curve)
		if err != nil {
			return "", "", err
		}
		desiredCR := scores[0].CR[poolName]
		newStarRating := findNewStarRating(firstPlaceWeighting, desiredCR)
		oldStarRating := leaderboardInfo.ForcedStarRating[poolName]
		new.WriteString(fmt.Sprintf("!set_manual %s %s %f\n", id, poolName, newStarRating))
		revert.WriteString(fmt.Sprintf("!set_manual %s %s %f\n", id, poolName, oldStarRating))
		loadingbar.Increment()
	}
	loadingbar.Finish()

	return new.String(), revert.String(), nil
}
