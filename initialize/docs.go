package initialize

import (
	"fmt"

	"go.uber.org/zap"
	"k8soperation/pkg/app"
)

func LogDocsReady(a *app.App) {
	base := fmt.Sprintf("http://127.0.0.1:%s", a.ServerSetting.Port)
	a.Logger.Info("docs ready",
		zap.String("swagger_ui", base+"/swagger/index.html"),
		zap.String("doc_json", base+"/swagger/doc.json"),
	)
}
