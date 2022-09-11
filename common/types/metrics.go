package types

// MetricsInfo tracks the metrics info of phalanx.
type MetricsInfo struct {
	// AvePackOrderLatency indicates interval since receive command to generate pre-order.
	AvePackOrderLatency float64

	//
	CurPackOrderLatency float64

	// AveOrderLatency indicates interval since receive command to generate partial order.
	AveOrderLatency float64

	//
	CurOrderLatency float64

	//
	AveOrderSize int

	// AveLogLatency indicates interval since generate partial order to commit partial order.
	AveLogLatency float64

	//
	CurLogLatency float64

	// AveCommandInfoLatency indicates since generate command info to commit command.
	AveCommandInfoLatency float64

	//
	CurCommandInfoLatency float64

	//
	AveCommitStreamLatency float64

	//
	CurCommitStreamLatency float64

	// SafeCommandCount indicates the number of command committed from safe path.
	SafeCommandCount int

	// RiskCommandCount indicates the number of command committed from risk path.
	RiskCommandCount int

	// FrontAttackFromRisk records the front attacked command requests from risk path.
	FrontAttackFromRisk int

	// FrontAttackFromSafe records the front attacked command requests from safe path.
	FrontAttackFromSafe int

	// FrontAttackIntervalRisk records the front attacked command requests of interval relationship from risk path.
	FrontAttackIntervalRisk int

	// FrontAttackIntervalSafe records the front attacked command requests of interval relationship from safe path.
	FrontAttackIntervalSafe int

	//
	SuccessRates []float64

	//=======
	// MSafeCommandCount indicates the number of command committed from safe path.
	MSafeCommandCount int

	// MRiskCommandCount indicates the number of command committed from risk path.
	MRiskCommandCount int

	// MFrontAttackFromRisk records the front attacked command requests from risk path.
	MFrontAttackFromRisk int

	// MFrontAttackFromSafe records the front attacked command requests from safe path.
	MFrontAttackFromSafe int

	// MFrontAttackIntervalRisk records the front attacked command requests of interval relationship from risk path.
	MFrontAttackIntervalRisk int

	// MFrontAttackIntervalSafe records the front attacked command requests of interval relationship from safe path.
	MFrontAttackIntervalSafe int

	//
	MSuccessRates []float64
	//=======

	//
	CommandPS float64

	//
	LogPS float64

	//
	GenLogPS float64
}
