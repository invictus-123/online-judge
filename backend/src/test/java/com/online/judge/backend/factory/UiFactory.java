package com.online.judge.backend.factory;

import com.online.judge.backend.dto.ui.ProblemDetailsUi;
import com.online.judge.backend.dto.ui.ProblemSummaryUi;
import com.online.judge.backend.dto.ui.SubmissionDetailsUi;
import com.online.judge.backend.dto.ui.SubmissionSummaryUi;
import com.online.judge.backend.dto.ui.TestCaseUi;
import com.online.judge.backend.dto.ui.UserSummaryUi;
import com.online.judge.backend.model.Problem;
import com.online.judge.backend.model.Submission;
import com.online.judge.backend.model.Tag;
import com.online.judge.backend.model.TestCase;
import com.online.judge.backend.model.User;
import com.online.judge.backend.model.shared.SolvedStatus;
import java.util.List;

public class UiFactory {

	public static ProblemDetailsUi createProblemDetailsUi(Problem problem) {
		return createProblemDetailsUi(problem, SolvedStatus.UNATTEMPTED);
	}

	public static ProblemDetailsUi createProblemDetailsUi(Problem problem, SolvedStatus solvedStatus) {
		return new ProblemDetailsUi(
				problem.getId(),
				problem.getTitle(),
				problem.getStatement(),
				problem.getTimeLimitSecond(),
				problem.getMemoryLimitMb(),
				problem.getDifficulty(),
				problem.getTags().stream().map(Tag::getTagName).toList(),
				problem.getTestCases().stream().map(tc -> createTestCaseUi(tc)).toList(),
				solvedStatus);
	}

	public static ProblemSummaryUi createProblemSummaryUi(Problem problem) {
		return createProblemSummaryUi(problem, SolvedStatus.UNATTEMPTED);
	}

	public static ProblemSummaryUi createProblemSummaryUi(Problem problem, SolvedStatus solvedStatus) {
		return new ProblemSummaryUi(
				problem.getId(),
				problem.getTitle(),
				problem.getDifficulty(),
				problem.getTags().stream().map(Tag::getTagName).toList(),
				solvedStatus);
	}

	public static SubmissionDetailsUi createSubmissionDetailsUi(Submission submission) {
		return new SubmissionDetailsUi(
				submission.getId(),
				createProblemSummaryUi(submission.getProblem()),
				createUserSummaryUi(submission.getUser()),
				submission.getStatus(),
				submission.getLanguage(),
				submission.getSubmittedAt(),
				submission.getCode(),
				submission.getExecutionTimeSeconds(),
				submission.getMemoryUsedMb(),
				List.of());
	}

	public static SubmissionSummaryUi createSubmissionSummaryUi(Submission submission) {
		return new SubmissionSummaryUi(
				submission.getId(),
				createProblemSummaryUi(submission.getProblem()),
				createUserSummaryUi(submission.getUser()),
				submission.getStatus(),
				submission.getLanguage(),
				submission.getSubmittedAt());
	}

	public static TestCaseUi createTestCaseUi(TestCase testCase) {
		return new TestCaseUi(testCase.getInput(), testCase.getOutput(), testCase.getExplanation());
	}

	public static UserSummaryUi createUserSummaryUi(User user) {
		return new UserSummaryUi(user.getHandle());
	}

	private UiFactory() {
		// Prevent instantiation
	}
}
