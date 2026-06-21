from rest_framework import serializers

from .models import Assessment, Question, QuestionOption, QuestionAsset, StudentAnswer, StudentSubmission
from .services import auto_mark_answer


class QuestionOptionSerializer(serializers.ModelSerializer):
    class Meta:
        model = QuestionOption
        fields = ["id", "question", "option_text", "is_correct", "order_number", "feedback"]


class QuestionAssetSerializer(serializers.ModelSerializer):
    class Meta:
        model = QuestionAsset
        fields = ["id", "question", "asset_type", "file", "caption", "alt_text", "created_at"]
        read_only_fields = ["created_at"]


class QuestionSerializer(serializers.ModelSerializer):
    options = QuestionOptionSerializer(many=True, read_only=True)
    assets = QuestionAssetSerializer(many=True, read_only=True)

    class Meta:
        model = Question
        fields = [
            "id", "assessment", "course", "created_by", "question_type", "question_text",
            "instruction", "marks", "order_number", "allow_whiteboard", "allow_image_upload",
            "allow_file_upload", "requires_manual_marking", "auto_marking_enabled", "metadata",
            "is_active", "options", "assets",
        ]
        read_only_fields = ["created_by", "requires_manual_marking", "auto_marking_enabled"]


class AssessmentSerializer(serializers.ModelSerializer):
    questions = QuestionSerializer(many=True, read_only=True)
    course_code = serializers.CharField(source="course.code", read_only=True)

    class Meta:
        model = Assessment
        fields = [
            "id", "course", "course_code", "title", "description", "assessment_type",
            "duration_minutes", "total_marks", "start_time", "end_time", "created_by", "status",
            "proctoring_level", "allow_mobile", "shuffle_questions", "shuffle_options",
            "show_result_immediately", "rules", "questions", "created_at", "updated_at",
        ]
        read_only_fields = ["created_by", "total_marks", "created_at", "updated_at"]


class StudentAnswerSerializer(serializers.ModelSerializer):
    question_detail = QuestionSerializer(source="question", read_only=True)

    class Meta:
        model = StudentAnswer
        fields = [
            "id", "submission", "question", "question_detail", "selected_option", "selected_option_ids",
            "text_answer", "blank_answers", "drag_drop_answer", "image_answer", "answer_file",
            "whiteboard_snapshot", "whiteboard_data", "is_auto_marked", "auto_score", "manual_score",
            "final_score", "marking_status", "lecturer_feedback", "marked_by", "marked_at", "submitted_at",
        ]
        read_only_fields = ["is_auto_marked", "auto_score", "manual_score", "final_score", "marking_status", "marked_by", "marked_at"]

    def create(self, validated_data):
        answer, _ = StudentAnswer.objects.update_or_create(
            submission=validated_data["submission"],
            question=validated_data["question"],
            defaults=validated_data,
        )
        return auto_mark_answer(answer)


class MarkAnswerSerializer(serializers.Serializer):
    manual_score = serializers.DecimalField(max_digits=7, decimal_places=2, min_value=0)
    lecturer_feedback = serializers.CharField(required=False, allow_blank=True)

    def validate_manual_score(self, value):
        answer = self.context["answer"]
        if value > answer.question.marks:
            raise serializers.ValidationError("Score cannot be higher than the question mark.")
        return value


class StudentSubmissionSerializer(serializers.ModelSerializer):
    answers = StudentAnswerSerializer(many=True, read_only=True)
    student_name = serializers.CharField(source="student.get_full_name", read_only=True)

    class Meta:
        model = StudentSubmission
        fields = ["id", "assessment", "student", "student_name", "status", "started_at", "submitted_at", "total_score", "released_at", "answers"]
        read_only_fields = ["student", "started_at", "submitted_at", "total_score", "released_at"]
