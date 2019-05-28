package ec2

import (
	"testing"
)

var region = "ap-northeast-1"

func TestListInstance(t *testing.T) {
	t.Log("List instance")

	cluster := ListInstance(nil, &region, "beepj-master")

	if cluster.Master.RawInstance == nil {
		t.Error("No found beej ")
	}
	t.Log("==>", cluster)
}
