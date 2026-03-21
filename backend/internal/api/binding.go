package api

import "github.com/gin-gonic/gin/binding"

func ConfigureBinding() {
	binding.EnableDecoderDisallowUnknownFields = true
}
