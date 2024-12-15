package main

import (
	"context"
	"encoding/json"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/invopop/jsonschema"
)

func main() {
	client := anthropic.NewClient()

	content := "Where is San Francisco?"

	println("[user]: " + content)

	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(content)),
	}

	tools := []anthropic.ToolParam{
		{
			Name:        anthropic.F("get_coordinates"),
			Description: anthropic.F("Accepts a place as an address, then returns the latitude and longitude coordinates."),
			InputSchema: anthropic.F(GetCoordinatesInputSchema),
		},
	}

	for {
		message, err := client.Messages.New(context.TODO(), anthropic.MessageNewParams{
			Model:     anthropic.F(anthropic.ModelClaude_3_5_Sonnet_20240620),
			MaxTokens: anthropic.Int(1024),
			Messages:  anthropic.F(messages),
			Tools:     anthropic.F(tools),
		})

		if err != nil {
			panic(err)
		}

		print("[assistant]: ")
		for _, block := range message.Content {
			switch block := block.AsUnion().(type) {
			case anthropic.TextBlock:
				println(block.Text)
			case anthropic.ToolUseBlock:
				println(block.Name + ": " + string(block.Input))
			}
		}

		messages = append(messages, message.ToParam())
		toolResults := []anthropic.ContentBlockParamUnion{}

		for _, block := range message.Content {
			if block.Type == anthropic.ContentBlockTypeToolUse {
				print("[user (" + block.Name + ")]: ")

				var response interface{}
				switch block.Name {
				case "get_coordinates":
					input := GetCoordinatesInput{}
					err := json.Unmarshal(block.Input, &input)
					if err != nil {
						panic(err)
					}
					response = GetCoordinates(input.Location)
				}

				b, err := json.Marshal(response)
				if err != nil {
					panic(err)
				}

				toolResults = append(toolResults, anthropic.NewToolResultBlock(block.ID, string(b), false))
			}
		}
		if len(toolResults) == 0 {
			break
		}
		messages = append(messages, anthropic.NewUserMessage(toolResults...))
	}
}

type GetCoordinatesInput struct {
	Location string `json:"location" jsonschema_description:"The location to look up."`
}

var GetCoordinatesInputSchema = GenerateSchema[GetCoordinatesInput]()

type GetCoordinateResponse struct {
	Long float64 `json:"long"`
	Lat  float64 `json:"lat"`
}

func GetCoordinates(location string) GetCoordinateResponse {
	//TODO: Implement a real API call to get the coordinates
	return GetCoordinateResponse{
		Long: -122.4194,
		Lat:  37.7749,
	}
}

func GenerateSchema[T any]() interface{} {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	return reflector.Reflect(v)
}
