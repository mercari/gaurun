package gcm

import "testing"

func TestValidateMessage(t *testing.T) {
	cases := []struct {
		input   *Message
		success bool
	}{
		{
			&Message{
				Data:            map[string]interface{}{"score": "5x1", "time": "15:10"},
				RegistrationIDs: []string{"4", "8", "15", "16", "23", "42"},
				TimeToLive:      600,
			},
			true,
		},

		{
			nil,
			false,
		},

		{
			&Message{},
			false,
		},

		{
			&Message{
				RegistrationIDs: []string{},
			},
			false,
		},

		// test should fail when more than 1000 RegistrationIDs are specifie
		{
			&Message{
				RegistrationIDs: make([]string, 1001),
			},
			false,
		},

		// test should fail when message TimeToLive field is negative
		{
			&Message{
				RegistrationIDs: []string{"1"},
				TimeToLive:      -1,
			},
			false,
		},

		// test should fail when message TimeToLive field is greater than 2419200
		{
			&Message{
				RegistrationIDs: []string{"1"},
				TimeToLive:      2419201,
			},
			false,
		},
	}

	for i, tc := range cases {
		err := tc.input.validate()
		if err != nil {
			if tc.success {
				t.Fatalf("#%d expect validate() to be failed", i)
			}
			continue
		}

		if !tc.success {
			t.Fatalf("#%d expect validate() not to be failed", i)
		}
	}
}
