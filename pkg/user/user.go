package user

import (
    "github.com/fatih/color"
    "strconv"
    "strings"
    "bufio"
    "context"
    "os"
    "time"
    "github.com/pranavbhat77/Quizz-Generator-GenAI/pkg/quiz"
)

var (
    green  = color.New(color.FgGreen, color.Bold)
    red    = color.New(color.FgRed, color.Bold)
    cyan   = color.New(color.FgCyan)
    yellow = color.New(color.FgYellow)
)

type UserScore struct {
    Name     string `json:"name"`
    Topic    string `json:"topic"`
    Score    int    `json:"score"`
    Attempts int    `json:"attempts"`
}

func HandleAnswer(answer int, q quiz.QuizQuestion, score *UserScore) {
    if answer < 1 || answer > 4 {
        red.Println("Invalid input. Skipping question.")
        return
    }

    score.Attempts++
    if answer-1 == q.Answer {
        green.Println("Correct! +4 points")
        score.Score += 4
    } else {
        red.Println("Wrong! -1 point")
        score.Score--
    }
}
