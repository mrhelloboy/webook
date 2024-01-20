package tencent

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"os"
	"testing"
)

func TestService_Send(t *testing.T) {
	secretId, ok := os.LookupEnv("SMS_SECRET_ID")
	if !ok {
		t.Fatal()
	}
	secretKey, ok := os.LookupEnv("SMS_SECRET_KEY")
	if !ok {
		t.Fatal()
	}
	appId, ok := os.LookupEnv("SMS_APP_ID")
	if !ok {
		t.Fatal()
	}
	signName, ok := os.LookupEnv("SMS_SIGN_NAME")
	if !ok {
		t.Fatal()
	}

	phoneNumb, ok := os.LookupEnv("SMS_PHONE_NUMBER")
	if !ok {
		t.Fatal()
	}

	c, err := sms.NewClient(common.NewCredential(secretId, secretKey), "ap-guangzhou", profile.NewClientProfile())
	if err != nil {
		t.Fatal(err)
	}
	s := NewService(c, appId, signName)
	testcase := []struct {
		name    string
		tplId   string
		params  []string
		numbers []string
		wantErr error
	}{
		{
			name:    "发送验证码",
			tplId:   "10000000",
			params:  []string{"123456"},
			numbers: []string{phoneNumb},
			wantErr: err,
		},
	}

	for _, tc := range testcase {
		t.Run(tc.name, func(t *testing.T) {
			er := s.Send(context.Background(), tc.tplId, tc.params, tc.numbers...)
			assert.Equal(t, tc.wantErr, er)
		})
	}
}
