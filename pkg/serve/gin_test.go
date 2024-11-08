package serve

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestGinDefault(t *testing.T) {
	type TestParams struct {
		Project string `json:"project" form:"project"`
		Model   string `json:"model" form:"model"`
		Number  int    `json:"number" form:"number"`
		Stream  bool   `json:"stream" form:"stream"`
	}
	router := gin.Default()
	router.GET("param", func(ctx *gin.Context) {
		params := TestParams{}
		if err := ctx.ShouldBindQuery(&params); err != nil {
			ctx.String(http.StatusBadRequest, "Invalid query param")
			// ctx.JSON(http.StatusBadRequest, gin.H{"code": 1001, "message": "Invalid query param"})
		}
		t.Log("Parse params", params)
		ctx.String(http.StatusOK, "ok")
	})

	go func() {
		router.Run(":9527")
	}()

	time.Sleep(2 * time.Second)

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := client.Get("http://127.0.0.1:9527/param?project=degpt&model=Qwen2.5-72B&number=20&stream=true")
	if err != nil {
		t.Fatalf("send http request error: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read http response body failed: %v", err)
	}
	t.Log("read http response", resp.Status, string(body))
}
