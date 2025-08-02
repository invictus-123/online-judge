package com.online.judge.backend.converter;

import com.online.judge.backend.dto.message.TestCaseResultMessage;
import com.online.judge.backend.model.Submission;
import com.online.judge.backend.model.TestCase;
import com.online.judge.backend.model.TestCaseResult;
import com.online.judge.backend.model.shared.TestCaseResultId;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.function.Function;
import java.util.stream.Collectors;

public class TestCaseResultConverter {
	public static TestCaseResult toTestCaseResult(
			Submission submission, TestCaseResultMessage testCaseResult, Map<UUID, TestCase> testCaseMap) {
		TestCaseResult result = new TestCaseResult();
		result.setId(new TestCaseResultId(submission.getId(), testCaseResult.testCaseId()));
		result.setSubmission(submission);
		result.setTestCase(testCaseMap.get(testCaseResult.testCaseId()));
		result.setVerdict(testCaseResult.status());
		result.setUserOutput(testCaseResult.output());
		result.setExecutionTimeSeconds(testCaseResult.timeTaken());
		result.setMemoryUsedMb(testCaseResult.memoryUsed());
		return result;
	}

	public static List<TestCaseResult> toTestCaseResult(
			Submission submission, List<TestCaseResultMessage> testCaseResults) {
		Map<UUID, TestCase> testCaseMap = submission.getProblem().getTestCases().stream()
				.collect(Collectors.toMap(TestCase::getId, Function.identity()));

		return testCaseResults.stream()
				.map(tcr -> toTestCaseResult(submission, tcr, testCaseMap))
				.toList();
	}

	private TestCaseResultConverter() {
		// Prevent instantiation
	}
}
