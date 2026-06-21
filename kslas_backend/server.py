import os

os.environ.setdefault("DJANGO_SETTINGS_MODULE", "kslas_backend.settings")

from django.core.handlers.wsgi import WSGIHandler

application = WSGIHandler()
