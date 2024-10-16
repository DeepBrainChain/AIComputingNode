package types

type ModelHistory struct {
	TimeStamp int64  `json:"timestamp"`
	ReqId     string `json:"req_id"`
	ReqNodeId string `json:"req_node_id"`
	ResNodeId string `json:"res_node_id"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Project   string `json:"project"`
	Model     string `json:"model"`
	// Chat Model Request
	ChatMessages []ChatCompletionMessage `json:"chat_messages"`
	// Chat Model Response
	ChatChoices []ChatResponseChoice `json:"chat_choices"`
	// Chat Model Response Usage
	ChatUsage ChatResponseUsage `json:"chat_usage"`
	// Image Generation Request
	ImagePrompt string `json:"image_prompt"`
	// Image Generation Response
	ImageChoices []ImageResponseChoice `json:"image_choices"`
}
