package com.online.judge.backend.service;

import static com.online.judge.backend.repository.specification.SubmissionSpecifications.and;
import static com.online.judge.backend.repository.specification.SubmissionSpecifications.hasProblem;
import static com.online.judge.backend.repository.specification.SubmissionSpecifications.hasProblemIdIn;
import static com.online.judge.backend.repository.specification.SubmissionSpecifications.hasStatusIn;
import static com.online.judge.backend.repository.specification.SubmissionSpecifications.hasUser;

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
	private final SubmissionRepository submissionRepository;

	public SolvedStatusService(SubmissionRepository submissionRepository) {
		this.submissionRepository = submissionRepository;
	}

	public SolvedStatus getSolvedStatus(User user, Long problemId) {
		if (user == null) {
			return SolvedStatus.UNATTEMPTED;
		}

		Specification<Submission> solvedSpec =
				and(hasUser(user), hasProblem(problemId), hasStatusIn(List.of(SubmissionStatus.PASSED)));
		if (submissionRepository.exists(solvedSpec)) {
			return SolvedStatus.SOLVED;
		}

		Specification<Submission> attemptedSpec = and(hasUser(user), hasProblem(problemId));
		if (submissionRepository.exists(attemptedSpec)) {
			return SolvedStatus.ATTEMPTED;
		}

		return SolvedStatus.UNATTEMPTED;
	}

	public Map<Long, SolvedStatus> getSolvedStatusForProblems(User user, List<Long> problemIds) {
		if (user == null) {
			return problemIds.stream().collect(Collectors.toMap(id -> id, id -> SolvedStatus.UNATTEMPTED));
		}

		Specification<Submission> solvedSpec =
				and(hasUser(user), hasProblemIdIn(problemIds), hasStatusIn(List.of(SubmissionStatus.PASSED)));
		Set<Long> solvedProblemIds = submissionRepository.findAll(solvedSpec).stream()
				.map(submission -> submission.getProblem().getId())
				.collect(Collectors.toSet());

		Specification<Submission> attemptedSpec = and(hasUser(user), hasProblemIdIn(problemIds));
		Set<Long> attemptedProblemIds = submissionRepository.findAll(attemptedSpec).stream()
				.map(submission -> submission.getProblem().getId())
				.collect(Collectors.toSet());

		return problemIds.stream().collect(Collectors.toMap(id -> id, id -> {
			if (solvedProblemIds.contains(id)) {
				return SolvedStatus.SOLVED;
			} else if (attemptedProblemIds.contains(id)) {
				return SolvedStatus.ATTEMPTED;
			} else {
				return SolvedStatus.UNATTEMPTED;
			}
		}));
	}
}
