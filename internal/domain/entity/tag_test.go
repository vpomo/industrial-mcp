package entity

import (
	"testing"
	"time"
)

func TestNewTag(t *testing.T) {
	tag, err := NewTag("temperature", 25.5)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tag.Name() != "temperature" {
		t.Errorf("expected name 'temperature', got %s", tag.Name())
	}
	if tag.Value() != 25.5 {
		t.Errorf("expected value 25.5, got %v", tag.Value())
	}
	if tag.Quality() != QualityGood {
		t.Errorf("expected quality QualityGood, got %v", tag.Quality())
	}
	if tag.ID() == "" {
		t.Error("expected non-empty ID")
	}
}

func TestNewTagEmptyName(t *testing.T) {
	_, err := NewTag("", 25.5)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestTagUpdateValue(t *testing.T) {
	tag, _ := NewTag("sensor", 10.0)
	oldTimestamp := tag.Timestamp()

	time.Sleep(10 * time.Millisecond)
	tag.UpdateValue(20.0)

	if tag.Value() != 20.0 {
		t.Errorf("expected value 20.0, got %v", tag.Value())
	}
	if !tag.Timestamp().After(oldTimestamp) {
		t.Error("expected timestamp to be updated")
	}
}

func TestTagSetQuality(t *testing.T) {
	tag, _ := NewTag("test", 100)
	tag.SetQuality(QualityBad)
	if tag.Quality() != QualityBad {
		t.Errorf("expected QualityBad, got %v", tag.Quality())
	}
}
