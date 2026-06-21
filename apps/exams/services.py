from decimal import Decimal

from .submission_models import StudentAnswer


def normalize_text(value):
    return str(value or "").strip().lower()


def auto_mark_answer(answer: StudentAnswer) -> StudentAnswer:
    question = answer.question

    if question.requires_manual_marking or not question.auto_marking_enabled:
        answer.marking_status = "needs_review"
        answer.is_auto_marked = False
        answer.final_score = Decimal("0")
        answer.save(update_fields=["marking_status", "is_auto_marked", "final_score"])
        return answer

    score = Decimal("0")

    if question.question_type == "single_choice":
        if answer.selected_option and answer.selected_option.is_correct:
            score = question.marks

    elif question.question_type == "multiple_choice":
        correct_ids = set(question.options.filter(is_correct=True).values_list("id", flat=True))
        selected_ids = set(int(item) for item in answer.selected_option_ids or [])
        if selected_ids == correct_ids:
            score = question.marks
        elif question.metadata.get("partial_marking") and correct_ids:
            correct_selected = len(selected_ids.intersection(correct_ids))
            wrong_selected = len(selected_ids.difference(correct_ids))
            ratio = max(0, correct_selected - wrong_selected) / len(correct_ids)
            score = question.marks * Decimal(str(ratio))

    elif question.question_type == "fill_blank":
        expected = [normalize_text(v) for v in question.metadata.get("correct_answers", [])]
        given = [normalize_text(v) for v in answer.blank_answers or []]
        if expected and given == expected:
            score = question.marks

    elif question.question_type == "drag_drop":
        expected = question.metadata.get("correct_order") or question.metadata.get("correct_answer")
        if expected and answer.drag_drop_answer == expected:
            score = question.marks

    answer.auto_score = score
    answer.final_score = score
    answer.is_auto_marked = True
    answer.marking_status = "auto_marked"
    answer.save(update_fields=["auto_score", "final_score", "is_auto_marked", "marking_status"])
    answer.submission.recalculate_score()
    return answer
