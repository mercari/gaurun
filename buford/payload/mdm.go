package payload

// MDM payload for mobile device management.
type MDM struct {
	Token string `json:"mdm"`
}

// Validate MDM payload.
func (p *MDM) Validate() error {
	if p == nil {
		return ErrIncomplete
	}

	// must have a token.
	if len(p.Token) == 0 {
		return ErrIncomplete
	}
	return nil
}
