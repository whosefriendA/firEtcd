package gateway

import (
	"encoding/json"
	"github.com/whosefriendA/firEtcd/common"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// 统一使用项目中的 Respond 结构体
	response := common.Respond{Code: status, Data: data}
	json.NewEncoder(w).Encode(response)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"msg": err.Error()})
}
