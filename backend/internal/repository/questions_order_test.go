package repository

import "testing"

func TestInsertQuestionIDPlacesQuestionAtRequestedPosition(t *testing.T) {
	ordered := insertQuestionID([]int64{10, 20, 30}, 40, 2)

	expected := []int64{10, 40, 20, 30}
	for idx, id := range expected {
		if ordered[idx] != id {
			t.Fatalf("expected order %v, got %v", expected, ordered)
		}
	}
}

func TestMoveQuestionIDReordersExistingQuestion(t *testing.T) {
	ordered, found := moveQuestionID([]int64{10, 20, 30, 40}, 20, 4)
	if !found {
		t.Fatalf("expected question to be found")
	}

	expected := []int64{10, 30, 40, 20}
	for idx, id := range expected {
		if ordered[idx] != id {
			t.Fatalf("expected order %v, got %v", expected, ordered)
		}
	}
}

func TestRemoveQuestionIDPreservesRemainingOrder(t *testing.T) {
	ordered, found := removeQuestionID([]int64{10, 20, 30}, 20)
	if !found {
		t.Fatalf("expected question to be found")
	}

	expected := []int64{10, 30}
	for idx, id := range expected {
		if ordered[idx] != id {
			t.Fatalf("expected order %v, got %v", expected, ordered)
		}
	}
}

func TestNormalizeQuestionPositionClampsRange(t *testing.T) {
	if position := normalizeQuestionPosition(0, 3); position != 3 {
		t.Fatalf("expected zero order to map to last position, got %d", position)
	}
	if position := normalizeQuestionPosition(10, 3); position != 3 {
		t.Fatalf("expected oversized order to clamp to last position, got %d", position)
	}
	if position := normalizeQuestionPosition(2, 3); position != 2 {
		t.Fatalf("expected exact order to stay unchanged, got %d", position)
	}
}
