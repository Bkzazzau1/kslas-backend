from django.conf import settings
from django.db import models


class Assessment(models.Model):
    ASSESSMENT_TYPES = (
        ("exam", "Exam"),
        ("graded_assessment", "Graded Assessment"),
        ("ungraded_assessment", "Ungraded Assessment"),
        ("practice", "Practice"),
    )
    STATUS_CHOICES = (
        ("draft", "Draft"),
        ("published", "Published"),
        ("active", "Active"),
        ("closed", "Closed"),
        ("archived", "Archived"),
    )

    course = models.ForeignKey("academics.Course", on_delete=models.PROTECT, related_name="assessments")
    title = models.CharField(max_length=240)
    description = models.TextField(blank=True)
    assessment_type = models.CharField(max_length=40, choices=ASSESSMENT_TYPES, default="graded_assessment")
    duration_minutes = models.PositiveIntegerField(default=30)
    total_marks = models.DecimalField(max_digits=7, decimal_places=2, default=0)
    start_time = models.DateTimeField(null=True, blank=True)
    end_time = models.DateTimeField(null=True, blank=True)
    created_by = models.ForeignKey(settings.AUTH_USER_MODEL, on_delete=models.PROTECT, related_name="created_assessments")
    status = models.CharField(max_length=30, choices=STATUS_CHOICES, default="draft")
    proctoring_level = models.CharField(max_length=20, default="none")
    allow_mobile = models.BooleanField(default=True)
    shuffle_questions = models.BooleanField(default=False)
    shuffle_options = models.BooleanField(default=False)
    show_result_immediately = models.BooleanField(default=False)
    rules = models.JSONField(default=dict, blank=True)
    created_at = models.DateTimeField(auto_now_add=True)
    updated_at = models.DateTimeField(auto_now=True)

    def __str__(self):
        return self.title

    def recalculate_total_marks(self):
        total = self.questions.aggregate(total=models.Sum("marks"))["total"] or 0
        self.total_marks = total
        self.save(update_fields=["total_marks", "updated_at"])


class Question(models.Model):
    QUESTION_TYPES = (
        ("single_choice", "Objective / Single Answer"),
        ("multiple_choice", "Multiple Answers"),
        ("essay", "Essay"),
        ("fill_blank", "Fill in the Blank"),
        ("drag_drop", "Drag and Drop"),
        ("image_question", "Image Question"),
    )

    assessment = models.ForeignKey(Assessment, on_delete=models.CASCADE, related_name="questions")
    course = models.ForeignKey("academics.Course", on_delete=models.PROTECT, related_name="questions")
    created_by = models.ForeignKey(settings.AUTH_USER_MODEL, on_delete=models.PROTECT, related_name="created_questions")
    question_type = models.CharField(max_length=40, choices=QUESTION_TYPES)
    question_text = models.TextField()
    instruction = models.TextField(blank=True)
    marks = models.DecimalField(max_digits=7, decimal_places=2, default=1)
    order_number = models.PositiveIntegerField(default=1)
    difficulty = models.CharField(max_length=20, default="medium")
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

    def __str__(self):
        return f"Q{self.order_number}"
