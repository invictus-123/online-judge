package com.online.judge.backend.factory;

import static com.online.judge.backend.factory.TagFactory.createTag;
import static com.online.judge.backend.factory.UserFactory.createUser;

import com.online.judge.backend.model.Problem;
import com.online.judge.backend.model.TestCase;
import com.online.judge.backend.model.User;
import com.online.judge.backend.model.shared.ProblemDifficulty;
import com.online.judge.backend.model.shared.ProblemTag;
import java.util.List;

public class ProblemFactory {
	private static final String TITLE = "Problem Title";
	private static final String STATEMENT = "Problem Statement";
	private static final Double TIME_LIMIT_SECOND = 1.0;
	private static final Integer MEMORY_LIMIT_MB = 1024;
	private static final ProblemDifficulty DIFFICULTY = ProblemDifficulty.EASY;
	private static final User CREATED_BY_USER = createUser();
	private static final List<ProblemTag> PROBLEM_TAGS = List.of(ProblemTag.ARRAY, ProblemTag.GREEDY);

	public static Problem createProblem() {
		return createProblem(
				TITLE, STATEMENT, TIME_LIMIT_SECOND, MEMORY_LIMIT_MB, DIFFICULTY, CREATED_BY_USER, PROBLEM_TAGS);
	}

	public static Problem createProblem(ProblemDifficulty difficulty, List<ProblemTag> problemTags) {
		return createProblem(
				TITLE, STATEMENT, TIME_LIMIT_SECOND, MEMORY_LIMIT_MB, difficulty, CREATED_BY_USER, problemTags);
	}

	public static Problem createProblem(
			String title,
			String statement,
			Double timeLimitSecond,
			Integer memoryLimitMb,
			ProblemDifficulty difficulty,
			User createdBy,
			List<ProblemTag> problemTags) {
		Problem problem = new Problem();
		problem.setTitle(title);
		problem.setStatement(statement);
		problem.setTimeLimitSecond(timeLimitSecond);
		problem.setMemoryLimitMb(memoryLimitMb);
		problem.setDifficulty(difficulty);
		problem.setCreatedBy(createdBy);
		problem.setTags(problemTags.stream()
				.map(problemTag -> createTag(problem, problemTag))
				.toList());
		return problem;
	}

	public static Problem createProblem(
			String title,
			String statement,
			Double timeLimitSecond,
			Integer memoryLimitMb,
			ProblemDifficulty difficulty,
			User createdBy,
			List<ProblemTag> problemTags,
			List<TestCase> testCases) {
		Problem problem =
				createProblem(title, statement, timeLimitSecond, memoryLimitMb, difficulty, createdBy, problemTags);
		problem.setTestCases(testCases);
		return problem;
	}

	private ProblemFactory() {
		// Prevent instantiation
	}
}
