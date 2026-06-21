from rest_framework import serializers
from .models import Course, Department


class DepartmentSerializer(serializers.ModelSerializer):
    class Meta:
        model = Department
        fields = ["id", "name", "code", "faculty", "is_active"]


class CourseSerializer(serializers.ModelSerializer):
    department_name = serializers.CharField(source="department.name", read_only=True)

    class Meta:
        model = Course
        fields = ["id", "code", "title", "department", "department_name", "level", "semester", "credit_units", "is_active"]
