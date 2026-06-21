from django.db import models


class Assessment(models.Model):
    title = models.CharField(max_length=240)
    description = models.TextField(blank=True)

    def __str__(self):
        return self.title
