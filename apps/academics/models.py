from django.db import models


class Department(models.Model):
    name = models.CharField(max_length=180)
    code = models.CharField(max_length=30, unique=True)
    faculty = models.CharField(max_length=180, blank=True)
    is_active = models.BooleanField(default=True)

    def __str__(self):
        return self.code


class Course(models.Model):
    code = models.CharField(max_length=40, unique=True)
    title = models.CharField(max_length=220)
    department = models.ForeignKey(Department, on_delete=models.PROTECT, related_name="courses")
    level = models.PositiveIntegerField(default=100)
    semester = models.CharField(max_length=20, default="first")
    credit_units = models.PositiveSmallIntegerField(default=3)
    is_active = models.BooleanField(default=True)

    def __str__(self):
        return self.code
