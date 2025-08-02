package com.online.judge.backend.dto.message;

import com.online.judge.backend.model.shared.SubmissionLanguage;
import jakarta.validation.constraints.NotNull;
import java.util.List;

/**
 * DTO representing to enqueue a message to be sent to the executor.
 */
public record SubmissionMessage(
		@NotNull Long submissionId,
		@NotNull SubmissionLanguage language,
		@NotNull String code,
		@NotNull Double timeLimit,
		@NotNull Integer memoryLimit,
		@NotNull List<TestCaseMessage> testCases) {}
