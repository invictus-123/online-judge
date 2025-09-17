package com.online.judge.executorv2.model.message;

import com.online.judge.executorv2.model.shared.SubmissionStatus;
import jakarta.validation.constraints.NotNull;

public record StatusUpdateMessage(@NotNull Long submissionId, @NotNull SubmissionStatus status) {}