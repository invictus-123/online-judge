package com.online.judge.backend.dto.message;

import com.online.judge.backend.model.shared.SubmissionStatus;
import jakarta.validation.constraints.NotNull;
import java.util.UUID;

/**
 * DTO representing result of a single test case.
 */
public record TestCaseResultMessage(
		@NotNull UUID testCaseId,
		@NotNull Double timeTaken,
		@NotNull Integer memoryUsed,
		@NotNull SubmissionStatus status,
		@NotNull String output) {}
