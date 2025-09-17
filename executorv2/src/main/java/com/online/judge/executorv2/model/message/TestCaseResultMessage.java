package com.online.judge.executorv2.model.message;

import com.online.judge.executorv2.model.shared.SubmissionStatus;
import jakarta.validation.constraints.NotNull;
import java.util.UUID;

public record TestCaseResultMessage(
		@NotNull UUID testCaseId,
		@NotNull Double timeTaken,
		@NotNull Integer memoryUsed,
		@NotNull SubmissionStatus status,
		@NotNull String output) {}