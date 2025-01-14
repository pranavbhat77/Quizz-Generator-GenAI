package utils

import (
    "io"
    "log"
    "os"

    "github.com/pdfcpu/pdfcpu/pkg/api"
    "github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
)

const maxPDFPages = 200

func ExtractTextFromPDF(file *os.File) (string, error) {
    ctx, err := api.ReadContext(file, pdfcpu.NewDefaultConfiguration())
    if err != nil {
        return "", err
    }

    if ctx.PageCount > maxPDFPages {
        return "", fmt.Errorf("PDF file exceeds the page limit of %d pages", maxPDFPages)
    }

    var text string
    for i := 1; i <= ctx.PageCount; i++ {
        pageText, err := api.ExtractPageContent(ctx, i)
        if err != nil {
            return "", err
        }
        text += pageText
    }

    return text, nil
}

func RemoveJSONFile(jsonFile string) {
    if _, err := os.Stat(jsonFile); err == nil {
        os.Remove(jsonFile)
    }
}
