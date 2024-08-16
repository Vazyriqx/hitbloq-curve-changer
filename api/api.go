package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"hitbloq-curve-changer/hitBloqtypes"
	"io"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

var rateLimiter = time.NewTicker(time.Millisecond * 500) // Allows 2 requests per second
const (
	maxRetries = 3
	retryDelay = time.Second * 5
)

func request(url string, target interface{}) error {
	for attempt := 1; attempt < maxRetries; attempt++ {

		<-rateLimiter.C

		log.Debugf("Fetching %s", url)
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		if string(body) == "[]" {
			if attempt < maxRetries {
				log.Warnf("Empty body received, retrying %s/%s times\n", attempt+1, maxRetries)
				time.Sleep(retryDelay)
				continue
			}
			return errors.New("response body is empty after retries")
		}
		if err := json.Unmarshal(body, &target); err != nil {
			return err
		}
		return nil

	}

	return nil

}

func GetLeaderBoardRankedList(poolName string, page int) (*hitBloqtypes.RankedList, error) {
	var rankedList hitBloqtypes.RankedList
	url := fmt.Sprintf("https://hitbloq.com/api/ranked_list/%s/%d", poolName, page)
	err := request(url, &rankedList)
	if err != nil {
		log.Errorf("No response from URL\n")
		return nil, err
	}

	return &rankedList, nil

}

func GetAllLeaderboardIDs(poolName string) ([]string, error) {
	var (
		leaderboardIds []string
		page           = 0
	)

	for {
		rankedList, err := GetLeaderBoardRankedList(poolName, page)
		if err != nil {
			return nil, err
		}
		leaderboardIds = append(leaderboardIds, rankedList.LeaderboardIDList...)
		if len(rankedList.LeaderboardIDList) != 30 {
			break
		}
		page++
	}

	return leaderboardIds, nil
}

func GetLeaderboardInfo(id string) (*hitBloqtypes.LeaderboardInfo, error) {
	url := fmt.Sprintf("https://hitbloq.com/api/leaderboard/%s/info", id)
	var leaderboardInfo hitBloqtypes.LeaderboardInfo
	if err := request(url, &leaderboardInfo); err != nil {
		return nil, err
	}
	return &leaderboardInfo, nil
}

func GetScores(id string, page int) ([]hitBloqtypes.Score, error) {
	url := fmt.Sprintf("https://hitbloq.com/api/leaderboard/%s/scores/%d", id, page)
	var scores []hitBloqtypes.Score
	if err := request(url, &scores); err != nil {
		return nil, err
	}
	return scores, nil
}
