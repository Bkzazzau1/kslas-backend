from rest_framework import viewsets
from .models import Course, Department
from .serializers import CourseSerializer, DepartmentSerializer


class DepartmentViewSet(viewsets.ModelViewSet):
    queryset = Department.objects.all()
    serializer_class = DepartmentSerializer


class CourseViewSet(viewsets.ModelViewSet):
    queryset = Course.objects.select_related("department").all()
    serializer_class = CourseSerializer
