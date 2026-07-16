package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"shop/pkg/proto"

	"github.com/anaskhan96/go-password-encoder"
)

func TestModelToResponse(t *testing.T) {
	bday := time.Date(2000, 1, 2, 3, 4, 5, 0, time.UTC)
	tests := []struct {
		name string
		user User
		want *proto.UserInfoResponse
	}{
		{
			name: "with birthday",
			user: User{
				Mobile:   "13800000000",
				Password: "hashed",
				NickName: "alice",
				Gender:   "female",
				Role:     1,
				Birthday: &bday,
			},
			want: &proto.UserInfoResponse{
				Id:       0,
				Password: "hashed",
				NickName: "alice",
				Gender:   "female",
				Role:     1,
				Mobile:   "13800000000",
				BirthDay: uint64(bday.Unix()),
			},
		},
		{
			name: "nil birthday leaves BirthDay zero",
			user: User{
				Mobile:   "13900000000",
				NickName: "bob",
				Gender:   "male",
				Role:     2,
				Birthday: nil,
			},
			want: &proto.UserInfoResponse{
				Password: "",
				NickName: "bob",
				Gender:   "male",
				Role:     2,
				Mobile:   "13900000000",
				BirthDay: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ModelToResponse(&tt.user)
			if got.Id != tt.want.Id ||
				got.Password != tt.want.Password ||
				got.NickName != tt.want.NickName ||
				got.Gender != tt.want.Gender ||
				got.Role != tt.want.Role ||
				got.Mobile != tt.want.Mobile ||
				got.BirthDay != tt.want.BirthDay {
				t.Fatalf("ModelToResponse = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestCheckPassWordRoundTrip(t *testing.T) {
	options := passwordOptions
	srv := &UserServer{}

	tests := []struct {
		name  string
		plain string
	}{
		{"ascii", "admin123"},
		{"with symbol", "p@ss w0rd!"},
		{"unicode", "复杂密码#42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			salt, encoded := password.Encode(tt.plain, options)
			encrypted := fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encoded)

			rsp, err := srv.CheckPassWord(context.Background(), &proto.PasswordCheckInfo{
				Password:          tt.plain,
				EncryptedPassword: encrypted,
			})
			if err != nil {
				t.Fatalf("CheckPassWord error: %v", err)
			}
			if !rsp.Success {
				t.Fatalf("expected password %q to verify", tt.plain)
			}
		})
	}
}

func TestCheckPassWordRejectsWrongPassword(t *testing.T) {
	options := passwordOptions
	salt, encoded := password.Encode("correct", options)
	encrypted := fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encoded)

	srv := &UserServer{}
	rsp, err := srv.CheckPassWord(context.Background(), &proto.PasswordCheckInfo{
		Password:          "wrong",
		EncryptedPassword: encrypted,
	})
	if err != nil {
		t.Fatalf("CheckPassWord error: %v", err)
	}
	if rsp.Success {
		t.Fatal("expected wrong password to fail verification")
	}
}
