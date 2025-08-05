package com.online.judge.backend.dto.ui;

import com.online.judge.backend.model.shared.ProblemDifficulty;
import com.online.judge.backend.model.shared.ProblemTag;
import com.online.judge.backend.model.shared.SolvedStatus;
import java.util.List;

public record ProblemSummaryUi(
		Long id, String title, ProblemDifficulty difficulty, List<ProblemTag> tags, SolvedStatus solvedStatus) {}
