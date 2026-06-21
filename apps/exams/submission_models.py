from django.conf import settings
from django.db import models
from django.utils import timezone

from .question_models import Question, QuestionOption


class QuestionAsset(models.Model):
    question = models.ForeignKey(Question, on_delete=models.CASCADE, related_name="assets")
    asset_type = models.CharField(max_length=20, default="image")
    file = models.FileField(upload_to="question_assets/")
    caption = models.CharField(max_length=255, blank=True)
    alt_text = models.CharField(max_length=255, blank=True)
    created_at = models.DateTimeField(auto_now_add=True)


class StudentSubmission(models.Model):
    STATUS_CHOICES = (
        ("in_progress", "In Progress"),
        ("submitted", "Submitted"),
        ("marked", "Marked"),
        ("released", "Released"),
    )
    assessment = models.ForeignKey("exams.Assessment", on_delete=models.PROTECT, related_name="submissions")
    student = models.ForeignKey(settings.AUTH_USER_MODEL, on_delete=models.PROTECT, related_name="assessment_submissions")
    status = models.CharField(max_length=30, choices=STATUS_CHOICES, default="in_progress")
    started_at = models.DateTimeField(default=timezone.now)
    submitted_at = models.DateTimeField(null=True, blank=True)
    total_score = models.DecimalField(max_digits=7, decimal_places=2, default=0)
    released_at = models.DateTimeField(null=True, blank=True)

    class Meta:
        unique_together = ("assessment", "student")
        ordering = ["-started_at"]

    def recalculate_score(self):
        total = self.answers.aggregate(total=models.Sum("final_score"))["total"] or 0
        self.total_score = total
        self.save(update_fields=["total_score"])


class StudentAnswer(models.Model):
    MARKING_CHOICES = (
        ("pending", "Pending"),
        ("auto_marked", "Auto Marked"),
        ("needs_review", "Needs Review"),
        ("marked", "Marked"),
    )
    submission = models.ForeignKey(StudentSubmission, on_delete=models.CASCADE, related_name="answers")
    question = models.ForeignKey(Question, on_delete=models.PROTECT, related_name="student_answers")
    selected_option = models.ForeignKey(QuestionOption, on_delete=models.SET_NULL, null=True, blank=True, related_name="single_choice_answers")
    selected_option_ids = models.JSONField(default=list, blank=True)
    text_answer = models.TextField(blank=True)
    blank_answers = models.JSONField(default=list, blank=True)
    drag_drop_answer = models.JSONField(default=dict, blank=True)
    image_answer = models.FileField(upload_to="student_image_answers/", blank=True, null=True)
    answer_file = models.FileField(upload_to="student_answer_files/", blank=True, null=True)
    whiteboard_snapshot = models.FileField(upload_to="whiteboard_snapshots/", blank=True, null=True)
    whiteboard_data = models.JSONField(default=dict, blank=True)
    is_auto_marked = models.BooleanField(default=False)
    auto_score = models.DecimalField(max_digits=7, decimal_places=2, default=0)
    manual_score = models.DecimalField(max_digits=7, decimal_places=2, null=True, blank=True)
    final_score = models.DecimalField(max_digits=7, decimal_places=2, default=0)
    marking_status = models.CharField(max_length=30, choices=MARKING_CHOICES, default="pending")
    lecturer_feedback = models.TextField(blank=True)
    marked_by = models.ForeignKey(settings.AUTH_USER_MODEL, on_delete=models.SET_NULL, null=True, blank=True, related_name="marked_answers")
    marked_at = models.DateTimeField(null=True, blank=True)
    submitted_at = models.DateTimeField(auto_now=True)

    class Meta:
        unique_together = ("submission", "question")
        ordering = ["question__order_number"]
