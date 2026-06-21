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
