package preprocessor

// ModularPipeline represents a pipeline composed of stages
type ModularPipeline struct {
	stages []PipelineStage
}

// NewModularPipeline creates a new modular pipeline
func NewModularPipeline(stages []PipelineStage) *ModularPipeline {
	return &ModularPipeline{
		stages: stages,
	}
}

// Execute runs all pipeline stages in sequence.
// The returned mapping maps positions in the final output back to original input offsets.
func (mp *ModularPipeline) Execute(input string, mapping []int) (string, []int, error) {
	result := input
	resultMapping := mapping

	for _, stage := range mp.stages {
		var err error
		result, resultMapping, err = stage.Execute(result, resultMapping)
		if err != nil {
			return result, resultMapping, err
		}
	}

	return result, resultMapping, nil
}

// StageCount returns the number of stages in the pipeline
func (mp *ModularPipeline) StageCount() int {
	return len(mp.stages)
}

// GetStageNames returns the names of all stages
func (mp *ModularPipeline) GetStageNames() []string {
	names := make([]string, len(mp.stages))
	for i, stage := range mp.stages {
		names[i] = stage.Name()
	}
	return names
}
