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

// Execute runs all pipeline stages in sequence
func (mp *ModularPipeline) Execute(input string) (string, error) {
	result := input

	for _, stage := range mp.stages {
		var err error
		result, err = stage.Execute(result)
		if err != nil {
			return result, err
		}
	}

	return result, nil
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
