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


class SubmissionViewSet(viewsets.ReadOnlyModelViewSet):
    queryset = StudentSubmission.objects.select_related("assessment", "student").prefetch_related("answers")
    serializer_class = StudentSubmissionSerializer

    @action(detail=True, methods=["post"])
    def release_result(self, request, pk=None):
        submission = self.get_object()
        submission.status = "released"
        submission.released_at = timezone.now()
        submission.save(update_fields=["status", "released_at"])
        return Response(self.get_serializer(submission).data)


class StudentAnswerViewSet(viewsets.ModelViewSet):
    queryset = StudentAnswer.objects.select_related("submission", "question", "selected_option")
    serializer_class = StudentAnswerSerializer

    @action(detail=True, methods=["patch"])
    def mark(self, request, pk=None):
        answer = self.get_object()
        serializer = MarkAnswerSerializer(data=request.data, context={"answer": answer})
        serializer.is_valid(raise_exception=True)
        answer.manual_score = serializer.validated_data["manual_score"]
        answer.final_score = serializer.validated_data["manual_score"]
        answer.lecturer_feedback = serializer.validated_data.get("lecturer_feedback", "")
        answer.marking_status = "marked"
        answer.marked_by = request.user
        answer.marked_at = timezone.now()
        answer.save(update_fields=["manual_score", "final_score", "lecturer_feedback", "marking_status", "marked_by", "marked_at"])
        answer.submission.recalculate_score()
        return Response(self.get_serializer(answer).data)


class StudentAssessmentViewSet(viewsets.ReadOnlyModelViewSet):
    queryset = Assessment.objects.filter(status__in=["published", "active"]).select_related("course").prefetch_related("questions")
    serializer_class = AssessmentSerializer

    @action(detail=True, methods=["post"])
    def start(self, request, pk=None):
        assessment = self.get_object()
        submission, _ = StudentSubmission.objects.get_or_create(assessment=assessment, student=request.user)
        return Response(StudentSubmissionSerializer(submission).data, status=status.HTTP_201_CREATED)

    @action(detail=True, methods=["post"])
    def submit(self, request, pk=None):
        assessment = self.get_object()
        submission = StudentSubmission.objects.get(assessment=assessment, student=request.user)
        submission.status = "submitted"
        submission.submitted_at = timezone.now()
        submission.save(update_fields=["status", "submitted_at"])
        return Response(StudentSubmissionSerializer(submission).data)
