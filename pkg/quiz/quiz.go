package quiz

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "strings"
    "time"

    "github.com/fatih/color"
    "github.com/google/generative-ai-go/genai"
    "google.golang.org/api/option"
)

type QuizQuestion struct {
    Question string   `json:"question"`
    Options  []string `json:"options"`
    Answer   int      `json:"answer"`
}

var (
    green  = color.New(color.FgGreen, color.Bold)
    red    = color.New(color.FgRed, color.Bold)
    cyan   = color.New(color.FgCyan)
    yellow = color.New(color.FgYellow)
)

const jsonFile = "quiz_data.json"

func GenerateQuestions(topic string) []QuizQuestion {
    ctx := context.Background()
    client, err := genai.NewClient(ctx, option.WithAPIKey(""))
    if err != nil {
        log.Fatalf("Error creating client: %v", err)
    }
    defer client.Close()

    model := client.GenerativeModel("gemini-1.5-flash-8b")
    model.SetTemperature(1)
    model.SetTopK(40)
    model.SetTopP(0.95)
    model.SetMaxOutputTokens(8192)

    prompt := fmt.Sprintf("Generate 5 latest quiz questions in JSON form for the topic %s with 4 options in this format:\n{\n    \"question\": \"Question text\",\n    \"options\": [\"option1\", \"option2\", \"option3\", \"option4\"],\n    \"answer\": correct_option_index\n}", topic)

    resp, err := model.GenerateContent(ctx, genai.Text(prompt))
    if err != nil {
        log.Fatalf("Error generating content: %v", err)
    }

    var questions []QuizQuestion
    var genaiText genai.Text
    if len(resp.Candidates) > 0 {
        for _, part := range resp.Candidates[0].Content.Parts {
            switch p := part.(type) {
            case genai.Text:
                genaiText = p
                genaiText = genai.Text(strings.ReplaceAll(strings.ReplaceAll(string(genaiText), "```", ""), "json", ""))
                err := json.Unmarshal([]byte(genaiText), &questions)
                if err != nil {
                    log.Fatalf("Error unmarshaling JSON: %v", err)
                }
            default:
                log.Printf("Unhandled part type: %T", p)
            }
        }
    } else {
        log.Fatalf("No candidates found in response")
    }
    return questions
}

func GenerateQuestionsFromText(text string) []QuizQuestion {
    ctx := context.Background()
    client, err := genai.NewClient(ctx, option.WithAPIKey(""))
    if err != nil {
        log.Fatalf("Error creating client: %v", err)
    }
    defer client.Close()

    model := client.GenerativeModel("gemini-1.5-flash-8b")
    model.SetTemperature(1)
    model.SetTopK(40)
    model.SetTopP(0.95)
    model.SetMaxOutputTokens(8192)

    prompt := fmt.Sprintf("Generate 5 latest quiz questions in JSON form from the following text with 4 options in this format:\n{\n    \"question\": \"Question text\",\n    \"options\": [\"option1\", \"option2\", \"option3\", \"option4\"],\n    \"answer\": correct_option_index\n}\n\nText:\n%s", text)

    resp, err := model.GenerateContent(ctx, genai.Text(prompt))
    if err != nil {
        log.Fatalf("Error generating content: %v", err)
    }

    var questions []QuizQuestion
    var genaiText genai.Text
    if len(resp.Candidates) > 0 {
        for _, part := range resp.Candidates[0].Content.Parts {
            switch p := part.(type) {
            case genai.Text:
                genaiText = p
                genaiText = genai.Text(strings.ReplaceAll(strings.ReplaceAll(string(genaiText), "```", ""), "json", ""))
                err := json.Unmarshal([]byte(genaiText), &questions)
                if err != nil {
                    log.Fatalf("Error unmarshaling JSON: %v", err)
                }
            default:
                log.Printf("Unhandled part type: %T", p)
            }
        }
    } else {
        log.Fatalf("No candidates found in response")
    }
    return questions
}

func SaveQuestionsToJSON(questions []QuizQuestion, jsonFile string) {
    existingQuestions := make([]QuizQuestion, 0)
    if _, err := os.Stat(jsonFile); err == nil {
        // File exists, read and append
        data, err := os.ReadFile(jsonFile)
        if err != nil {
            log.Fatalf("Error reading existing file: %v", err)
        }
        err = json.Unmarshal(data, &existingQuestions)
        if err != nil {
            log.Fatalf("Error unmarshaling existing JSON: %v", err)
        }
        // Append new questions
        existingQuestions = append(existingQuestions, questions...)
    } else {
        // File doesn't exist, just write the new questions
        existingQuestions = questions
    }

    file, err := json.MarshalIndent(existingQuestions, "", "    ")
    if err != nil {
        log.Fatalf("Error marshaling JSON: %v", err)
    }

    err = os.WriteFile(jsonFile, file, 0644)
    if err != nil {
        log.Fatalf("Error writing file: %v", err)
    }
}

func ConductQuiz(name string, topic string, questions []QuizQuestion) user.UserScore {
    score := user.UserScore{
        Name:  name,
        Topic: topic,
    }

    for i, q := range questions {
        // Display question and options first
        yellow.Printf("\nQuestion %d: %s\n", i+1, q.Question)
        for j, opt := range q.Options {
            cyan.Printf("%d. %s\n", j+1, opt)
        }
        cyan.Print("\nEnter your answer (1-4): ")

        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        answerCh := make(chan int)
        errorCh := make(chan error)

        go func() {
            reader := bufio.NewReader(os.Stdin)
            input, err := reader.ReadString('\n')
            if err != nil {
                errorCh <- err
                return
            }
            answer, err := strconv.Atoi(strings.TrimSpace(input))
            if err != nil {
                errorCh <- err
                return
            }
            answerCh <- answer
        }()

        // Handle timeout and answer
        select {
        case <-ctx.Done():
            red.Println("\nTime's up! Moving to next question")
            score.Attempts++
            score.Score--
        case err := <-errorCh:
            red.Printf("\nError reading input: %v\n", err)
        case answer := <-answerCh:
            user.HandleAnswer(answer, q, &score)
        }
    }
    return score
}
