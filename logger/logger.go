package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

var (
	Info    *log.Logger
	Error   *log.Logger
	Warning *log.Logger
	Debug   *log.Logger
)

// InitLogger инициализирует систему логирования
func InitLogger(logFilePath string) error {
	// Создаем директорию для логов, если ее нет
	logDir := filepath.Dir(logFilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("не удалось создать директорию для логов: %v", err)
	}

	// Открываем файл логов с опцией добавления в конец
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл логов: %v", err)
	}

	// Инициализируем логгеры
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	Info = log.New(multiWriter, "INFO: ", log.Ldate|log.Ltime)
	Warning = log.New(multiWriter, "WARNING: ", log.Ldate|log.Ltime)
	Error = log.New(multiWriter, "ERROR: ", log.Ldate|log.Ltime)
	Debug = log.New(multiWriter, "DEBUG: ", log.Ldate|log.Ltime)

	Info.Println("Система логирования инициализирована")
	return nil
}

// LogInfo выводит информационное сообщение с указанием источника
func LogInfo(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	Info.Printf("[%s:%d] %s", filepath.Base(file), line, fmt.Sprintf(format, v...))
}

// LogError выводит сообщение об ошибке с указанием источника
func LogError(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	Error.Printf("[%s:%d] %s", filepath.Base(file), line, fmt.Sprintf(format, v...))
}

// LogWarning выводит предупреждение с указанием источника
func LogWarning(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	Warning.Printf("[%s:%d] %s", filepath.Base(file), line, fmt.Sprintf(format, v...))
}

// LogDebug выводит отладочное сообщение с указанием источника
func LogDebug(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(1)
	Debug.Printf("[%s:%d] %s", filepath.Base(file), line, fmt.Sprintf(format, v...))
}
