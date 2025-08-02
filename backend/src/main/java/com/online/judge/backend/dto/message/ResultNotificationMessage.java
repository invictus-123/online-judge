package com.online.judge.backend.dto.message;

import com.online.judge.backend.model.shared.SubmissionStatus;
import jakarta.validation.constraints.NotNull;
import java.util.List;

/**
 * DTO representing result of the execution of a submission.
 */
public record ResultNotificationMessage(
		@NotNull Long submissionId,
		@NotNull SubmissionStatus status,
		@NotNull Double timeTaken,
		@NotNull Integer memoryUsed,
		List<TestCaseResultMessage> testCaseResults) {}
