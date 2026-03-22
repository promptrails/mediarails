// Package media provides a registry of all media providers and a convenience
// constructor for creating providers by name.
//
// # Usage
//
//	provider, err := media.New(media.ElevenLabs, "api-key")
//	// or
//	provider := media.MustNew(media.ElevenLabs, "api-key")
//
//	resp, err := provider.Generate(ctx, &mediarails.GenerateRequest{
//		Type:   mediarails.TTS,
//		Model:  "eleven_multilingual_v2",
//		Prompt: "Hello world!",
//	})
//
// All providers are registered automatically. No additional imports needed.
package media
