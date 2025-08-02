package com.online.judge.backend.dto.message;

import com.online.judge.backend.model.shared.SubmissionStatus;
import jakarta.validation.constraints.NotNull;

/**
 * DTO representing the update in the status of a submission.
 */
public record StatusUpdateMessage(@NotNull Long submissionId, @NotNull SubmissionStatus status) {}
