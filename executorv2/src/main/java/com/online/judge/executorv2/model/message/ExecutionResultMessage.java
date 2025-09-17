package com.online.judge.executorv2.model.message;

import com.online.judge.executorv2.model.shared.SubmissionStatus;
import jakarta.validation.constraints.NotNull;
import java.util.List;

public record ExecutionResultMessage(
		@NotNull Long submissionId,
		@NotNull SubmissionStatus status,
		@NotNull Double timeTaken,
		@NotNull Integer memoryUsed,
		List<TestCaseResultMessage> testCaseResults) {}