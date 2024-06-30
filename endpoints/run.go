package endpoints

import (
	"dafny-server/run"
	"encoding/json"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

type RequestBody2 struct {
	Requester string `json:"requester"`
	Files     []run.DafnyFile
}

func HandleRun(c run.RunService) func(ctx echo.Context) error {
	return func(ctx echo.Context) error {
		rb2 := RequestBody2{}
		data, err := io.ReadAll(ctx.Request().Body)
		if err != nil {
			return ctx.String(http.StatusBadRequest, "")
		}

		err = json.Unmarshal(data, &rb2)
		if err != nil {
			return ctx.String(http.StatusBadRequest, "Invalid JSON")
		}

		resultChan := make(chan run.RunResult)
		c.AddCodeInstanceToQueue(run.CodeInstance{
			Requester: rb2.Requester,
			Files:     rb2.Files,
			Result:    resultChan,
		})

		res := <-resultChan
		return ctx.String(res.Status, res.Content)

	}
}
