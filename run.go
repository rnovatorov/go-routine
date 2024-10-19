package routine

import "context"

type Run func(context.Context) error
