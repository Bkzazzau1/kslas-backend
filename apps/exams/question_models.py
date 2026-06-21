from django.conf import settings
from django.db import models


class Question(models.Model):
    QUESTION_TYPES = (
        ("single_choice", "Objective / Single Answer"),
        ("multiple_choice", "Multiple Answers"),
        ("essay", "Essay"),
        ("fill_blank", "Fill in the Blank"),
        ("drag_drop", "Drag and Drop"),
        ("image_question", "Image Question"),
    )
    assessment = models.ForeignKey("exams.Assessment", on_delete=models.CASCADE, related_name="questions")
    course = models.ForeignKey("academics.Course", on_delete=models.PROTECT, related_name="questions")
    created_by = models.ForeignKey(settings.AUTH_USER_MODEL, on_delete=models.PROTECT, related_name="created_questions")
    question_type = models.CharField(max_length=40, choices=QUESTION_TYPES)
    question_text = models.TextField()
    instruction = models.TextField(blank=True)
    marks = models.DecimalField(max_digits=7, decimal_places=2, default=1)
    order_number = models.PositiveIntegerField(default=1)
    allow_whiteboard = models.BooleanField(default=False)
    allow_image_upload = models.BooleanField(default=False)
    allow_file_upload = models.BooleanField(default=False)
    requires_manual_marking = models.BooleanField(default=False)
    auto_marking_enabled = models.BooleanField(default=True)
    metadata = models.JSONField(default=dict, blank=True)
    is_active = models.BooleanField(default=True)

    class Meta:
        ordering = ["assessment", "order_number", "id"]

    def save(self, *args, **kwargs):
        if self.question_type in ["essay", "image_question"] or self.allow_whiteboard or self.allow_image_upload or self.allow_file_upload:
            self.requires_manual_marking = True
            self.auto_marking_enabled = False
        super().save(*args, **kwargs)
        self.assessment.recalculate_total_marks()


class QuestionOption(models.Model):
    question = models.ForeignKey(Question, on_delete=models.CASCADE, related_name="options")
    option_text = models.TextField(blank=True)
    is_correct = models.BooleanField(default=False)
    order_number = models.PositiveIntegerField(default=1)
    feedback = models.TextField(blank=True)

    class Meta:
        ordering = ["order_number", "id"]
