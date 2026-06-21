from django.conf import settings
from django.conf.urls.static import static
from django.contrib import admin
from django.urls import include, path
from rest_framework.routers import DefaultRouter

from apps.academics.views import CourseViewSet, DepartmentViewSet
from apps.exams.views import LecturerAssessmentViewSet, LecturerQuestionViewSet, QuestionAssetViewSet, QuestionOptionViewSet, StudentAnswerViewSet, StudentAssessmentViewSet, SubmissionViewSet

router = DefaultRouter()
router.register("academics/departments", DepartmentViewSet, basename="department")
router.register("academics/courses", CourseViewSet, basename="course")
router.register("lecturer/assessments", LecturerAssessmentViewSet, basename="lecturer-assessment")
router.register("lecturer/questions", LecturerQuestionViewSet, basename="lecturer-question")
router.register("lecturer/options", QuestionOptionViewSet, basename="lecturer-question-option")
router.register("lecturer/assets", QuestionAssetViewSet, basename="lecturer-question-asset")
router.register("lecturer/submissions", SubmissionViewSet, basename="lecturer-submission")
router.register("lecturer/answers", StudentAnswerViewSet, basename="lecturer-answer")
router.register("student/assessments", StudentAssessmentViewSet, basename="student-assessment")

urlpatterns = [
    path("admin/", admin.site.urls),
    path("api/", include(router.urls)),
]

if settings.DEBUG:
    urlpatterns += static(settings.MEDIA_URL, document_root=settings.MEDIA_ROOT)
