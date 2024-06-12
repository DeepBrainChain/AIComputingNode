package model

import "testing"

func TestExecuteModel(t *testing.T) {
	var (
		api    = "http://127.0.0.1:8080"
		model  = "superimage"
		prompt = "bird"
	)
	code, message, image := ImageGenerationModel(api, model, prompt)
	if code != 0 {
		t.Fatalf("Execute model %s with %q error %s", model, prompt, message)
	}
	t.Logf("Execute model %s with %q result %v", model, prompt, image)
}
