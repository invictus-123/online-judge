package com.online.judge.backend.service;

import static com.online.judge.backend.repository.attributes.ProblemAttributes.ID;
import static com.online.judge.backend.repository.attributes.SubmissionAttributes.PROBLEM;
import static com.online.judge.backend.repository.attributes.SubmissionAttributes.STATUS;
import static com.online.judge.backend.repository.attributes.SubmissionAttributes.USER;
import static com.online.judge.backend.repository.specification.BaseSpecifications.and;
import static com.online.judge.backend.repository.specification.BaseSpecifications.hasAttributeInValues;
import static com.online.judge.backend.repository.specification.BaseSpecifications.hasAttributeWithValue;
import static com.online.judge.backend.repository.specification.BaseSpecifications.hasNestedAttributeInValues;
import static com.online.judge.backend.repository.specification.BaseSpecifications.hasNestedAttributeWithValue;

import com.online.judge.backend.model.Submission;
import com.online.judge.backend.model.User;
import com.online.judge.backend.model.shared.SolvedStatus;
import com.online.judge.backend.model.shared.SubmissionStatus;
import com.online.judge.backend.repository.SubmissionRepository;
import java.util.List;
import java.util.Map;
import java.util.Set;
import java.util.stream.Collectors;
import org.springframework.data.jpa.domain.Specification;
import org.springframework.stereotype.Service;

@Service
public class SolvedStatusService {
	private static final List<SubmissionStatus> FAILED_STATUSES = List.of(
			SubmissionStatus.TIME_LIMIT_EXCEEDED,
			SubmissionStatus.MEMORY_LIMIT_EXCEEDED,
			SubmissionStatus.COMPILATION_ERROR,
			SubmissionStatus.RUNTIME_ERROR);
	private final SubmissionRepository submissionRepository;

	public SolvedStatusService(SubmissionRepository submissionRepository) {
		this.submissionRepository = submissionRepository;
	}

	public SolvedStatus getSolvedStatus(User user, Long problemId) {
		return getSolvedStatusForProblems(user, List.of(problemId)).getOrDefault(problemId, SolvedStatus.UNATTEMPTED);
	}

	public Map<Long, SolvedStatus> getSolvedStatusForProblems(User user, List<Long> problemIds) {
		if (user == null) {
			return problemIds.stream().collect(Collectors.toMap(id -> id, id -> SolvedStatus.UNATTEMPTED));
		}

		Specification<Submission> solvedSpec = and(
				hasAttributeWithValue(USER, user),
				hasNestedAttributeInValues(PROBLEM, ID, problemIds),
				hasAttributeInValues(STATUS, List.of(SubmissionStatus.PASSED)));
		Set<Long> solvedProblemIds = submissionRepository.findAll(solvedSpec).stream()
				.map(submission -> submission.getProblem().getId())
				.collect(Collectors.toSet());

		Specification<Submission> failedAttemptSpec = and(
				hasAttributeWithValue(USER, user),
				hasNestedAttributeInValues(PROBLEM, ID, problemIds),
				hasAttributeInValues(STATUS, FAILED_STATUSES));
		Set<Long> failedAttemptProblemIds = submissionRepository.findAll(failedAttemptSpec).stream()
				.map(submission -> submission.getProblem().getId())
				.collect(Collectors.toSet());

		return problemIds.stream().collect(Collectors.toMap(id -> id, id -> {
			if (solvedProblemIds.contains(id)) {
				return SolvedStatus.SOLVED;
			} else if (failedAttemptProblemIds.contains(id)) {
				return SolvedStatus.FAILED_ATTEMPT;
			} else {
				return SolvedStatus.UNATTEMPTED;
			}
		}));
	}
}
