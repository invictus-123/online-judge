package com.online.judge.backend.dto.ui;

import com.online.judge.backend.model.shared.ProblemDifficulty;
import com.online.judge.backend.model.shared.ProblemTag;
import com.online.judge.backend.model.shared.SolvedStatus;
import java.util.List;

public record ProblemDetailsUi(
		Long id,
		String title,
		String statement,
		Double timeLimitInSecond,
		Integer memoryLimitInMb,
		ProblemDifficulty difficulty,
		List<ProblemTag> tags,
		List<TestCaseUi> sampleTestCases,
		SolvedStatus solvedStatus) {}
