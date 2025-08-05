package com.online.judge.backend.converter;

import static com.online.judge.backend.converter.TagConverter.toTagFromProblemTag;
import static com.online.judge.backend.converter.TestCaseConverter.toTestCaseFromCreateTestCaseRequest;
import static com.online.judge.backend.converter.TestCaseConverter.toTestCaseUi;

import com.online.judge.backend.dto.request.CreateProblemRequest;
import com.online.judge.backend.dto.ui.ProblemDetailsUi;
import com.online.judge.backend.dto.ui.ProblemSummaryUi;
import com.online.judge.backend.model.Problem;
import com.online.judge.backend.model.Tag;
import com.online.judge.backend.model.TestCase;
import com.online.judge.backend.model.shared.ProblemTag;
import com.online.judge.backend.model.shared.SolvedStatus;
import java.util.List;

public class ProblemConverter {
	public static ProblemSummaryUi toProblemSummaryUi(Problem problem) {
		return toProblemSummaryUi(problem, SolvedStatus.UNATTEMPTED);
	}

	public static ProblemSummaryUi toProblemSummaryUi(Problem problem, SolvedStatus solvedStatus) {
		return new ProblemSummaryUi(
				problem.getId(), problem.getTitle(), problem.getDifficulty(), listProblemTags(problem), solvedStatus);
	}

	public static ProblemDetailsUi toProblemDetailsUi(Problem problem) {
		return toProblemDetailsUi(problem, SolvedStatus.UNATTEMPTED);
	}

	public static ProblemDetailsUi toProblemDetailsUi(Problem problem, SolvedStatus solvedStatus) {
		return new ProblemDetailsUi(
				problem.getId(),
				problem.getTitle(),
				problem.getStatement(),
				problem.getTimeLimitSecond(),
				problem.getMemoryLimitMb(),
				problem.getDifficulty(),
				listProblemTags(problem),
				toTestCaseUi(listSampleTestCases(problem)),
				solvedStatus);
	}

	public static Problem toProblemFromCreateProblemRequest(CreateProblemRequest request) {
		Problem problem = new Problem();
		problem.setTitle(request.title());
		problem.setStatement(request.statement());
		problem.setDifficulty(request.difficulty());
		problem.setTimeLimitSecond(request.timeLimit());
		problem.setMemoryLimitMb(request.memoryLimit());

		List<Tag> tags = request.tags().stream()
				.map(problemTag -> toTagFromProblemTag(problem, problemTag))
				.toList();
		problem.setTags(tags);

		List<TestCase> testCases = request.testCases().stream()
				.map(tcRequest -> toTestCaseFromCreateTestCaseRequest(problem, tcRequest))
				.toList();
		problem.setTestCases(testCases);

		return problem;
	}

	private static List<ProblemTag> listProblemTags(Problem problem) {
		return problem.getTags().stream().map(Tag::getTagName).toList();
	}

	private static List<TestCase> listSampleTestCases(Problem problem) {
		return problem.getTestCases().stream().filter(TestCase::getIsSample).toList();
	}

	private ProblemConverter() {
		// Prevent instantiation
	}
}
