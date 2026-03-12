// Package mediarails provides a unified interface for AI media generation in Go.
//
// It supports speech (TTS/STT), image generation, and video generation across
// 8 providers through a single Provider interface. Both synchronous and
// asynchronous generation patterns are supported.
//
// # Quick Start
//
//	provider := elevenlabs.New("api-key")
//	resp, _ := provider.Generate(ctx, &mediarails.GenerateRequest{
//		Type:   mediarails.TTS,
//		Model:  "eleven_multilingual_v2",
//		Prompt: "Hello world!",
//		Config: map[string]any{"voice_id": "21m00Tcm4TlvDq8ikWAM"},
//	})
//	// resp.AssetData contains audio/mpeg bytes
package mediarails
