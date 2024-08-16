# Hitbloq Curve Changer
## Overview
simple CLI tool to change the curve of a pool while preserving the top CR play in the pool

## usage 

```
./hitbloq_curve_changer <poolname> <new-curve-json>
```

### exmaples
```hitbloq-curve-changer poodles '{"type": "basic", "baseline": 0.78, "cutoff": 0.5, "exponential": 2.5}'```

```hitbloq-curve-changer poodles '{"type": "linear", "points": [[0.0, 0.0], [0.8, 0.5], [1.0, 1.0]]}'```

# Contributing
I am open to pull requests. 

# Contact
`Vazyriqx` on discord. Easiest place to find me is in the hitbloq discord