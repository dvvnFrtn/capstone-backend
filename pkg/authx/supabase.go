package authx

import (
	"os"

	"github.com/supabase-community/auth-go"
)

func NewSupabase() auth.Client {
	return auth.New(os.Getenv("SUPABASE_PROJECT_REFERENCE"), os.Getenv("SUPABASE_ANON_KEY"))
}
