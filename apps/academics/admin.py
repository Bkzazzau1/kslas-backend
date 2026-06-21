from django.contrib import admin
from .models import Course, Department


@admin.register(Department)
class DepartmentAdmin(admin.ModelAdmin):
    list_display = ("code", "name", "faculty", "is_active")
    search_fields = ("code", "name", "faculty")


@admin.register(Course)
class CourseAdmin(admin.ModelAdmin):
    list_display = ("code", "title", "department", "level", "semester", "credit_units", "is_active")
    search_fields = ("code", "title")
    list_filter = ("department", "level", "semester", "is_active")
