package com.online.judge.backend.service;

import static com.online.judge.backend.factory.ProblemFactory.createProblem;
import static com.online.judge.backend.factory.SubmissionFactory.createSubmission;
import static com.online.judge.backend.factory.UserFactory.createUser;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.mockito.Mockito.when;

import com.online.judge.backend.model.Problem;
import com.online.judge.backend.model.Submission;
import com.online.judge.backend.model.User;
import com.online.judge.backend.model.shared.SolvedStatus;
import com.online.judge.backend.model.shared.SubmissionLanguage;
import com.online.judge.backend.model.shared.SubmissionStatus;
import com.online.judge.backend.repository.SubmissionRepository;
import java.util.List;
import java.util.Map;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentMatchers;
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
		Problem problem = createProblem();
		problem.setId(problemId);
		Submission solvedSubmission =
				createSubmission(user, problem, SubmissionStatus.PASSED, SubmissionLanguage.JAVA, "code", 1.0, 512);

		when(submissionRepository.findAll(ArgumentMatchers.<Specification<Submission>>any()))
				.thenReturn(List.of(solvedSubmission)) // First call: solved problems
				.thenReturn(List.of()); // Second call: failed attempts

		SolvedStatus result = solvedStatusService.getSolvedStatus(user, problemId);
		assertEquals(SolvedStatus.SOLVED, result);
	}

	@Test
	void getSolvedStatus_whenUserHasFailedAttemptButNotSolved_returnsFailedAttempt() {
		User user = createUser();
		Long problemId = 1L;
		Problem problem = createProblem();
		problem.setId(problemId);
		Submission failedSubmission = createSubmission(
				user, problem, SubmissionStatus.RUNTIME_ERROR, SubmissionLanguage.JAVA, "code", 1.0, 512);

		when(submissionRepository.findAll(ArgumentMatchers.<Specification<Submission>>any()))
				.thenReturn(List.of()) // First call: no solved problems
				.thenReturn(List.of(failedSubmission)); // Second call: failed attempts

		SolvedStatus result = solvedStatusService.getSolvedStatus(user, problemId);
		assertEquals(SolvedStatus.FAILED_ATTEMPT, result);
	}

	@Test
	void getSolvedStatus_whenUserHasNeverSubmitted_returnsUnattempted() {
		User user = createUser();
		Long problemId = 1L;

		when(submissionRepository.findAll(ArgumentMatchers.<Specification<Submission>>any()))
				.thenReturn(List.of()) // First call: no solved problems
				.thenReturn(List.of()); // Second call: no failed attempts

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
		Problem problem1 = createProblem();
		problem1.setId(1L);
		Problem problem2 = createProblem();
		problem2.setId(2L);
		Submission solvedSubmission =
				createSubmission(user, problem1, SubmissionStatus.PASSED, SubmissionLanguage.JAVA, "code", 1.0, 512);
		Submission attemptedSubmission = createSubmission(
				user, problem2, SubmissionStatus.RUNTIME_ERROR, SubmissionLanguage.JAVA, "code", 1.0, 512);
		when(submissionRepository.findAll(ArgumentMatchers.<Specification<Submission>>any()))
				.thenReturn(List.of(solvedSubmission)) // First call: solved problems
				.thenReturn(List.of(solvedSubmission, attemptedSubmission)); // Second call: attempted problems

		Map<Long, SolvedStatus> result = solvedStatusService.getSolvedStatusForProblems(user, problemIds);

		assertEquals(3, result.size());
		assertEquals(SolvedStatus.SOLVED, result.get(1L));
		assertEquals(SolvedStatus.FAILED_ATTEMPT, result.get(2L));
		assertEquals(SolvedStatus.UNATTEMPTED, result.get(3L));
	}
}
