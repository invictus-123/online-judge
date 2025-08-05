package com.online.judge.backend.service;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;
import static com.online.judge.backend.factory.UserFactory.createUser;

import com.online.judge.backend.model.Problem;
import com.online.judge.backend.model.Submission;
import com.online.judge.backend.model.User;
import com.online.judge.backend.model.shared.SolvedStatus;
import com.online.judge.backend.model.shared.SubmissionStatus;
import com.online.judge.backend.repository.SubmissionRepository;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.data.jpa.domain.Specification;

@ExtendWith(MockitoExtension.class)
class SolvedStatusServiceTest {

	@Mock
	private SubmissionRepository submissionRepository;

	private SolvedStatusService solvedStatusService;

	@BeforeEach
	void setUp() {
		solvedStatusService = new SolvedStatusService(submissionRepository);
	}

	@Test
	void getSolvedStatus_whenUserIsNull_returnsUnattempted() {
		SolvedStatus result = solvedStatusService.getSolvedStatus(null, 1L);
		assertEquals(SolvedStatus.UNATTEMPTED, result);
	}

	@Test
	void getSolvedStatus_whenUserHasSolvedProblem_returnsSolved() {
		User user = createUser();
		Long problemId = 1L;

		when(submissionRepository.exists(any(Specification.class))).thenReturn(true);

		SolvedStatus result = solvedStatusService.getSolvedStatus(user, problemId);
		assertEquals(SolvedStatus.SOLVED, result);
	}

	@Test
	void getSolvedStatus_whenUserHasAttemptedButNotSolved_returnsAttempted() {
		User user = createUser();
		Long problemId = 1L;

		when(submissionRepository.exists(any(Specification.class)))
				.thenReturn(false) // Not solved
				.thenReturn(true); // But attempted

		SolvedStatus result = solvedStatusService.getSolvedStatus(user, problemId);
		assertEquals(SolvedStatus.ATTEMPTED, result);
	}

	@Test
	void getSolvedStatus_whenUserHasNeverSubmitted_returnsUnattempted() {
		User user = createUser();
		Long problemId = 1L;

		when(submissionRepository.exists(any(Specification.class)))
				.thenReturn(false) // Not solved
				.thenReturn(false); // Not attempted

		SolvedStatus result = solvedStatusService.getSolvedStatus(user, problemId);
		assertEquals(SolvedStatus.UNATTEMPTED, result);
	}

	@Test
	void getSolvedStatusForProblems_whenUserIsNull_returnsAllUnattempted() {
		List<Long> problemIds = List.of(1L, 2L, 3L);

		Map<Long, SolvedStatus> result = solvedStatusService.getSolvedStatusForProblems(null, problemIds);

		assertEquals(3, result.size());
		assertEquals(SolvedStatus.UNATTEMPTED, result.get(1L));
		assertEquals(SolvedStatus.UNATTEMPTED, result.get(2L));
		assertEquals(SolvedStatus.UNATTEMPTED, result.get(3L));
	}

	@Test
	void getSolvedStatusForProblems_withMixedStatuses_returnsCorrectStatuses() {
		User user = createUser();
		List<Long> problemIds = List.of(1L, 2L, 3L);
		Problem problem1 = new Problem();
		problem1.setId(1L);
		Problem problem2 = new Problem();
		problem2.setId(2L);
		Submission solvedSubmission = new Submission();
		solvedSubmission.setProblem(problem1);
		solvedSubmission.setUser(user);
		solvedSubmission.setStatus(SubmissionStatus.PASSED);
		Submission attemptedSubmission = new Submission();
		attemptedSubmission.setProblem(problem2);
		attemptedSubmission.setUser(user);
		attemptedSubmission.setStatus(SubmissionStatus.RUNTIME_ERROR);
		when(submissionRepository.findAll(any(Specification.class)))
				.thenReturn(List.of(solvedSubmission)) // First call: solved problems
				.thenReturn(List.of(solvedSubmission, attemptedSubmission)); // Second call: attempted problems

		Map<Long, SolvedStatus> result = solvedStatusService.getSolvedStatusForProblems(user, problemIds);

		assertEquals(3, result.size());
		assertEquals(SolvedStatus.SOLVED, result.get(1L));
		assertEquals(SolvedStatus.ATTEMPTED, result.get(2L));
		assertEquals(SolvedStatus.UNATTEMPTED, result.get(3L));
	}
}
