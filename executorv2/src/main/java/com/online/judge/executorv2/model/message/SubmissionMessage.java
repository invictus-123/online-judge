package com.online.judge.executorv2.model.message;

import com.online.judge.executorv2.model.shared.SubmissionLanguage;
import jakarta.validation.constraints.NotNull;
import java.util.List;

public record SubmissionMessage(
		@NotNull Long submissionId,
		@NotNull SubmissionLanguage language,
		@NotNull String code,
		@NotNull Double timeLimit,
		@NotNull Integer memoryLimit,
		@NotNull List<TestCaseMessage> testCases) {}