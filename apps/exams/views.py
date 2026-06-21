from django.utils import timezone
from rest_framework import status, viewsets
from rest_framework.decorators import action
from rest_framework.response import Response

from .models import Assessment, Question, QuestionOption, QuestionAsset, StudentAnswer, StudentSubmission
from .serializers import AssessmentSerializer, MarkAnswerSerializer, QuestionAssetSerializer, QuestionOptionSerializer, QuestionSerializer, StudentAnswerSerializer, StudentSubmissionSerializer


class LecturerAssessmentViewSet(viewsets.ModelViewSet):
    queryset = Assessment.objects.select_related("course", "created_by").prefetch_related("questions")
    serializer_class = AssessmentSerializer

    def perform_create(self, serializer):
        serializer.save(created_by=self.request.user)

    @action(detail=True, methods=["post"])
    def publish(self, request, pk=None):
        assessment = self.get_object()
        assessment.status = "published"
        assessment.save(update_fields=["status", "updated_at"])
        return Response(self.get_serializer(assessment).data)

    @action(detail=True, methods=["post"])
    def close(self, request, pk=None):
        assessment = self.get_object()
        assessment.status = "closed"
        assessment.save(update_fields=["status", "updated_at"])
        return Response(self.get_serializer(assessment).data)


class LecturerQuestionViewSet(viewsets.ModelViewSet):
    queryset = Question.objects.select_related("assessment", "course", "created_by").prefetch_related("options", "assets")
    serializer_class = QuestionSerializer

    def perform_create(self, serializer):
        assessment = serializer.validated_data["assessment"]
        serializer.save(created_by=self.request.user, course=assessment.course)


class QuestionOptionViewSet(viewsets.ModelViewSet):
    queryset = QuestionOption.objects.select_related("question")
    serializer_class = QuestionOptionSerializer


class QuestionAssetViewSet(viewsets.ModelViewSet):
    queryset = QuestionAsset.objects.select_related("question")
    serializer_class = QuestionAssetSerializer
