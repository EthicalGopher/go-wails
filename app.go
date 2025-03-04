package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"github.com/google/generative-ai-go/genai"
	"github.com/pinecone-io/go-pinecone/pinecone"
	"google.golang.org/api/option"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

const filepath = "temp/history.txt"
const apiFilepath = "temp/apt.txt"
const GeminiAPI = "AIzaSyDEFDl6zTZf0YaxtLQJ5Yu-SRXZ85bamm8"

var PineconeAPI = Readapidata()

const key = "1234567890123456"
const Api = "AIzaSyASVvG4dtZtpdqF2yEOZZtYK4hRxjBNxcA"

var _, _ = os.Create(filepath)

func EncryptAES(plainText, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	nonce := make([]byte, 12) // 12 bytes nonce for GCM
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	cipherText := aesGCM.Seal(nonce, nonce, []byte(plainText), nil)
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// DecryptAES decrypts a ciphertext message using AES-GCM.
func DecryptAES(cipherText, key string) (string, error) {
	if cipherText == "" {
		return "", fmt.Errorf("error: cipherText is empty")
	}

	data, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", fmt.Errorf("Base64 Decode Error: %v", err)
	}

	if len(data) < 12 { // ✅ Ensure data is at least 12 bytes
		return "", fmt.Errorf("error: invalid ciphertext, too short")
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce, cipherData := data[:12], data[12:]
	plainText, err := aesGCM.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %v", err)
	}

	return string(plainText), nil
}

func (a *App) CollectAPI(input string) string {
	if input == "" {
		return "No input provided to save."
	}

	

	newData := input
	if err := os.WriteFile(apiFilepath, []byte(newData), 0644); err != nil {
		log.Printf("Error writing to api file: %v", err)
		return "Failed to save your input."
	}

	return "Input saved successfully!"
}

func Readapidata() string {
	data, err := os.ReadFile(apiFilepath)
	if err != nil {
		fmt.Println(err)
	}
	return string(data)
}

func (a *App) Botresponse(userMessage string) string {

	if userMessage == "" {
		return "empty text"
	}

	data, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Println(err)

		data, _ = os.ReadFile(filepath)

	}

	var botResponse string

	botResponse, err = Response1(filepath, userMessage, string(data))

	if err != nil {
		fmt.Println(err)

	}
	return botResponse
}

func Response1(filepath, input, prev string) (string, error) { // Return error
	context1 := Queryvector(input)

	API := Api
	prev, _ = DecryptAES(prev, key)

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(API))
	if err != nil {
		return "", fmt.Errorf("creating client: %v", err) // Return error
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.0-flash") // **VERIFY MODEL NAME**
	model.SetTemperature(2)
	cs := model.StartChat()
	info := `
    
  You are MindfulAI, an empathetic AI-powered mental health assistant designed to provide real-time emotional support and guidance with Indian cultural sensitivity. Your primary role is to actively listen, analyze emotional tone, and respond with compassionate, tailored advice.

You engage users in meaningful, supportive conversations, offering:
✅ Guided conversations to help users express and process their emotions.
✅ Practical coping strategies for stress, anxiety, and emotional challenges.
✅ Self-help resources that empower users on their mental well-being journey.
✅ Interactive mindfulness exercises, structured using HTML tags like <ol><li> for clarity and engagement.

Key Guidelines:
✅ Use Context Window Only When Requested: If the user explicitly asks for information from the context window (e.g., "Can you suggest some meditation exercises?"), then use it for retrieval. Otherwise, do not refer to it.
✅ Confidentiality & Privacy: Ensure a safe, judgment-free space where users feel comfortable sharing.
✅ Empathetic & Solution-Oriented: Maintain kindness, cultural awareness, and constructive guidance in all responses.
✅ Stay Focused: Do not discuss unrelated topics like coding, mathematics, or technical subjects.

Your Mission:
Your goal is to provide immediate, safe, and empathetic support, guiding users toward emotional well-being through insightful, structured, and engaging conversations.  
    
    `
	cs.History = []*genai.Content{
		{
			Parts: []genai.Part{
				genai.Text(info),
			},
			Role: "model",
		},
	}

	query := `
    
    Context = ` + context1 + `
	Previouse chat =` + prev + `
    
    User Input = ` + input + `
    
    
    
    
    `

	res, err := cs.SendMessage(ctx, genai.Text(query))
	if err != nil {
		return "", fmt.Errorf("sending message: %v", err)
	}

	var botResponse string
	for _, cand := range res.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				botResponse += fmt.Sprintf("%v", part)
			}
		}
	}
	encryptedreply, _ := EncryptAES(botResponse, key)

	os.WriteFile(filepath, []byte(encryptedreply), 0644)

	return botResponse, nil
}

func Queryvector(data string) string {
	vector := Embedding(data)
	Searched := SearchData(vector)
	content := ""
	for _, j := range Searched {
		content += j
	}
	return content
}

func SearchData(queryVector []float32) []string {
	indexName := "mindfulai"
	namespace := "any"
	field := "text"
	API := PineconeAPI
	ctx := context.Background()

	clientParams := pinecone.NewClientParams{
		ApiKey: API,
	}
	pc, err := pinecone.NewClient(clientParams)
	if err != nil {
		log.Println(err)
	}
	filepathxModel, err := pc.DescribeIndex(ctx, indexName)
	if err != nil {
		log.Fatalf("Failed to describe index \"%v\": %v", indexName, err)
	}

	filepathxConnection, err := pc.Index(pinecone.NewIndexConnParams{Host: filepathxModel.Host, Namespace: namespace})
	if err != nil {
		log.Fatalf("Failed to create IndexConnection1 for Host %v: %v", filepathxModel.Host, err)
	}

	res, err := filepathxConnection.QueryByVectorValues(ctx, &pinecone.QueryByVectorValuesRequest{
		Vector:          queryVector,
		TopK:            3,
		IncludeValues:   true,
		IncludeMetadata: true,
	})
	var values []string
	if err != nil {
		log.Fatalf("Error encountered when querying by vector: %v", err)
	} else {
		for _, match := range res.Matches {

			value := fmt.Sprintln(match.Vector.Metadata.Fields[field])
			values = append(values, value)
		}
	}
	return values
}

func Embedding(data string) []float32 {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, option.WithAPIKey(GeminiAPI))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	em := client.EmbeddingModel("text-embedding-004")
	res, err := em.EmbedContent(ctx, genai.Text(data))

	if err != nil {
		panic(err)
	}
	return []float32(res.Embedding.Values)
}
