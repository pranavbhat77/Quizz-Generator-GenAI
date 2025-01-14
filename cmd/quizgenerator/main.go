package main

import (
    "bufio"
    "context"
    "encoding/json"
    "fmt"
    "log"
    "mime/multipart"
    "net/http"
    "os"
    "os/signal"
    "strconv"
    "strings"
    "syscall"
    "time"

    "github.com/fatih/color"
    "github.com/google/generative-ai-go/genai"
    "google.golang.org/api/option"
    "github.com/pranavbhat77/Quizz-Generator-GenAI/pkg/quiz"
    "github.com/pranavbhat77/Quizz-Generator-GenAI/pkg/user"
    "github.com/pranavbhat77/Quizz-Generator-GenAI/pkg/utils"
)

const jsonFile = "quiz_data.json"
const maxPDFSize = 10 * 1024 * 1024 // 10 MB
const maxPDFPages = 200

var (
    green  = color.New(color.FgGreen, color.Bold)
    red    = color.New(color.FgRed, color.Bold)
    cyan   = color.New(color.FgCyan)
    yellow = color.New(color.FgYellow)
)

func main() {
    cleanup := make(chan os.Signal, 1)
    signal.Notify(cleanup, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-cleanup
        utils.RemoveJSONFile(jsonFile)
        os.Exit(0)
    }()

    reader := bufio.NewReader(os.Stdin)
    cyan.Print("Enter your name: ")
    name, _ := reader.ReadString('\n')
    name = strings.TrimSpace(name)

    cyan.Print("Enter the topic for quiz: ")
    topic, _ := reader.ReadString('\n')
    topic = strings.TrimSpace(topic)

    cyan.Print("Do you want to upload a PDF for questions? (yes/no): ")
    uploadPDF, _ := reader.ReadString('\n')
    uploadPDF = strings.TrimSpace(uploadPDF)

    var questions []quiz.QuizQuestion
    if strings.ToLower(uploadPDF) == "yes" {
        questions = handlePDFUpload()
    } else {
        questions = quiz.GenerateQuestions(topic)
    }

    quiz.SaveQuestionsToJSON(questions, jsonFile)
    score := quiz.ConductQuiz(name, topic, questions)

    yellow.Printf("\nFinal Results for %s:\n", name)
    yellow.Printf("Topic: %s\n", topic)
    if score.Score > 0 {
        green.Printf("Score: %d\n", score.Score)
    } else {
        red.Printf("Score: %d\n", score.Score)
    }
    yellow.Printf("Total Attempts: %d\n", score.Attempts)

    utils.RemoveJSONFile(jsonFile)
}

func handlePDFUpload() []quiz.QuizQuestion {
    cyan.Print("Enter the path to the PDF file: ")
    reader := bufio.NewReader(os.Stdin)
    pdfPath, _ := reader.ReadString('\n')
    pdfPath = strings.TrimSpace(pdfPath)

    file, err := os.Open(pdfPath)
    if err != nil {
        log.Fatalf("Error opening PDF file: %v", err)
    }
    defer file.Close()

    fileInfo, err := file.Stat()
    if err != nil {
        log.Fatalf("Error getting file info: %v", err)
    }

    if fileInfo.Size() > maxPDFSize {
        log.Fatalf("PDF file size exceeds the limit of %d MB", maxPDFSize/(1024*1024))
    }

    text, err := utils.ExtractTextFromPDF(file)
    if err != nil {
        log.Fatalf("Error extracting text from PDF: %v", err)
    }

    return quiz.GenerateQuestionsFromText(text)
}
