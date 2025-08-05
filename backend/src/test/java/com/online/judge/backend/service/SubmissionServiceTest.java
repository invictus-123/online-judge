package com.online.judge.backend.service;

import static com.online.judge.backend.factory.ProblemFactory.createProblem;
import static com.online.judge.backend.factory.SubmissionFactory.createSubmission;
import static com.online.judge.backend.factory.UiFactory.createSubmissionDetailsUi;
import static com.online.judge.backend.factory.UserFactory.createUser;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.verifyNoInteractions;
import static org.mockito.Mockito.when;

import com.online.judge.backend.dto.filter.SubmissionFilterRequest;
import com.online.judge.backend.dto.request.SubmitCodeRequest;
import com.online.judge.backend.dto.ui.SubmissionDetailsUi;
import com.online.judge.backend.dto.ui.SubmissionSummaryUi;
import com.online.judge.backend.exception.ProblemNotFoundException;
import com.online.judge.backend.exception.SubmissionNotFoundException;
import com.online.judge.backend.model.Problem;
import com.online.judge.backend.model.Submission;
import com.online.judge.backend.model.User;
import com.online.judge.backend.model.shared.SubmissionLanguage;
import com.online.judge.backend.model.shared.SubmissionStatus;
import com.online.judge.backend.queue.SubmissionPublisher;
import com.online.judge.backend.repository.ProblemRepository;
import com.online.judge.backend.repository.SubmissionRepository;
import com.online.judge.backend.util.UserUtil;
import java.util.List;
import java.util.Optional;
import java.util.stream.Stream;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.Arguments;
import org.junit.jupiter.params.provider.MethodSource;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageImpl;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.Sort;
import org.springframework.data.jpa.domain.Specification;

@ExtendWith(MockitoExtension.class)
class SubmissionServiceTest {
	private static final int PAGE_SIZE = 20;

	@Mock
	private ProblemRepository problemRepository;

	@Mock
	private SubmissionRepository submissionRepository;

	@Mock
	private SubmissionPublisher submissionPublisher;

	@Mock
	private UserUtil userUtil;

	private SubmissionService submissionService;

	@BeforeEach
	void setUp() {
		submissionService = new SubmissionService(
				problemRepository, submissionRepository, submissionPublisher, userUtil, PAGE_SIZE);
	}

	static Stream<Arguments> listSubmissionsScenarios() {
		return Stream.of(
				// Basic scenarios - no filters
				Arguments.of("no filters - submissions exist", false, null, null, null, 1, 2, false, false),
				Arguments.of("no filters - no submissions exist", false, null, null, null, 1, 0, true, false),

				// Only me filter
				Arguments.of("only me filter - with authenticated user", true, null, null, null, 1, 1, false, true),
				Arguments.of("only me filter - no authenticated user", true, null, null, null, 1, 2, false, false),

				// Problem ID filter
				Arguments.of("problem ID filter", false, 1L, null, null, 1, 1, false, false),
				Arguments.of("problem ID filter - no matches", false, 99L, null, null, 1, 0, true, false),

				// Status filtering
				Arguments.of(
						"single status filter",
						false,
						null,
						List.of(SubmissionStatus.PASSED),
						null,
						1,
						1,
						false,
						false),
				Arguments.of(
						"multiple statuses filter",
						false,
						null,
						List.of(SubmissionStatus.PASSED, SubmissionStatus.RUNTIME_ERROR),
						null,
						1,
						2,
						false,
						false),
				Arguments.of(
						"status filter - no matches",
						false,
						null,
						List.of(SubmissionStatus.COMPILATION_ERROR),
						null,
						1,
						0,
						true,
						false),

				// Language filtering
				Arguments.of(
						"single language filter",
						false,
						null,
						null,
						List.of(SubmissionLanguage.JAVA),
						1,
						1,
						false,
						false),
				Arguments.of(
						"multiple languages filter",
						false,
						null,
						null,
						List.of(SubmissionLanguage.JAVA, SubmissionLanguage.PYTHON),
						1,
						2,
						false,
						false),
				Arguments.of(
						"language filter - no matches",
						false,
						null,
						null,
						List.of(SubmissionLanguage.CPP),
						1,
						0,
						true,
						false),

				// Combined filtering
				Arguments.of(
						"combined problem and status filtering",
						false,
						1L,
						List.of(SubmissionStatus.PASSED),
						null,
						1,
						1,
						false,
						false),
				Arguments.of(
						"combined status and language filtering",
						false,
						null,
						List.of(SubmissionStatus.PASSED),
						List.of(SubmissionLanguage.JAVA),
						1,
						1,
						false,
						false),
				Arguments.of(
						"only me with problem filter - authenticated user", true, 1L, null, null, 1, 1, false, true),

				// Pagination
				Arguments.of("custom page number", false, null, null, null, 3, 1, false, false),
				Arguments.of("page 5 with results", false, null, null, null, 5, 2, false, false));
	}

	@ParameterizedTest(name = "listSubmissions - {0}")
	@MethodSource("listSubmissionsScenarios")
	void listSubmissions_allScenarios(
			String scenario,
			Boolean onlyMe,
			Long problemId,
			List<SubmissionStatus> statuses,
			List<SubmissionLanguage> languages,
			int page,
			int expectedResultCount,
			boolean isEmpty,
			boolean hasAuthenticatedUser) {

		SubmissionFilterRequest filterRequest =
				new SubmissionFilterRequest(onlyMe, problemId, statuses, languages, page);

		// Mock authenticated user if needed
		User authenticatedUser = null;
		if (hasAuthenticatedUser) {
			authenticatedUser = createUser("testuser");
			when(userUtil.getCurrentAuthenticatedUser()).thenReturn(authenticatedUser);
		} else if (Boolean.TRUE.equals(onlyMe)) {
			// Simulate no authenticated user case
			when(userUtil.getCurrentAuthenticatedUser()).thenThrow(new RuntimeException("No authenticated user"));
		}

		// Create appropriate mock submissions based on scenario
		List<Submission> submissions;
		if (isEmpty) {
			submissions = List.of();
		} else if (expectedResultCount == 1) {
			Submission submission = createSubmissionWithId(1L);
			if (hasAuthenticatedUser) {
				submission.setUser(authenticatedUser);
			}
			if (problemId != null) {
				Problem problem = createProblem();
				problem.setId(problemId);
				submission.setProblem(problem);
			}
			if (statuses != null && !statuses.isEmpty()) {
				submission.setStatus(statuses.get(0));
			}
			if (languages != null && !languages.isEmpty()) {
				submission.setLanguage(languages.get(0));
			}
			submissions = List.of(submission);
		} else if (expectedResultCount == 2) {
			Submission submission1 = createSubmissionWithId(1L);
			Submission submission2 = createSubmissionWithId(2L);

			if (statuses != null && statuses.size() >= 2) {
				submission1.setStatus(statuses.get(0));
				submission2.setStatus(statuses.get(1));
			}
			if (languages != null && languages.size() >= 2) {
				submission1.setLanguage(languages.get(0));
				submission2.setLanguage(languages.get(1));
			}
			submissions = List.of(submission1, submission2);
		} else {
			submissions = List.of(createSubmissionWithId(1L));
		}

		Pageable expectedPageable =
				PageRequest.of(page - 1, PAGE_SIZE, Sort.by("submittedAt").descending());
		Page<Submission> submissionPage =
				isEmpty ? Page.empty() : new PageImpl<>(submissions, expectedPageable, submissions.size());
		when(submissionRepository.findAll(any(Specification.class), any(Pageable.class)))
				.thenReturn(submissionPage);

		List<SubmissionSummaryUi> result = submissionService.listSubmissions(filterRequest);

		// Assertions
		assertNotNull(result);
		assertEquals(expectedResultCount, result.size());
		verify(submissionRepository).findAll(any(Specification.class), any(Pageable.class));
	}

	@Test
	void getSubmissionDetailsById_whenSubmissionExists_shouldReturnDetails() {
		Long submissionId = 10L;
		Submission submission = createSubmissionWithId(submissionId);
		when(submissionRepository.findById(submissionId)).thenReturn(Optional.of(submission));
		SubmissionDetailsUi expectedSubmmissionDetails = createSubmissionDetailsUi(submission);

		SubmissionDetailsUi result = submissionService.getSubmissionDetailsById(submissionId);

		assertNotNull(result);
		assertEquals(expectedSubmmissionDetails, result);
	}

	@Test
	void getSubmissionDetailsById_whenSubmissionDoesNotExist_shouldThrowException() {
		Long submissionId = 10L;
		when(submissionRepository.findById(submissionId)).thenReturn(Optional.empty());

		SubmissionNotFoundException exception = assertThrows(SubmissionNotFoundException.class, () -> {
			submissionService.getSubmissionDetailsById(submissionId);
		});

		assertEquals("Submission with ID " + submissionId + " not found", exception.getMessage());
	}

	@Test
	void submitCode_whenProblemExists_shouldCreateAndReturnSubmissionId() {
		long problemId = 1L;
		long submissionId = 123L;
		SubmitCodeRequest request = new SubmitCodeRequest(problemId, "public class Main {}", SubmissionLanguage.JAVA);
		User user = createUser();
		Problem problem = createProblem();
		problem.setId(problemId);
		Submission submission = createSubmission();
		submission.setId(submissionId);
		submission.setUser(user);
		submission.setProblem(problem);
		SubmissionDetailsUi expectedSubmissionDetails = createSubmissionDetailsUi(submission);
		when(userUtil.getCurrentAuthenticatedUser()).thenReturn(user);
		when(problemRepository.findById(problemId)).thenReturn(Optional.of(problem));
		when(submissionRepository.save(any(Submission.class))).thenReturn(submission);

		SubmissionDetailsUi submissionDetails = submissionService.submitCode(request);

		assertEquals(expectedSubmissionDetails, submissionDetails);
		ArgumentCaptor<Submission> submissionCaptor = ArgumentCaptor.forClass(Submission.class);
		verify(submissionRepository).save(submissionCaptor.capture());
		Submission savedSubmission = submissionCaptor.getValue();
		assertEquals(user, savedSubmission.getUser());
		assertEquals(problem, savedSubmission.getProblem());
		assertEquals(request.code(), savedSubmission.getCode());
		assertEquals(request.language(), savedSubmission.getLanguage());
		assertEquals(SubmissionStatus.WAITING_FOR_EXECUTION, savedSubmission.getStatus());
		verify(submissionPublisher).sendSubmission(any());
	}

	@Test
	void submitCode_whenProblemNotFound_shouldThrowException() {
		long problemId = 99L;
		SubmitCodeRequest request = new SubmitCodeRequest(problemId, "code", SubmissionLanguage.PYTHON);
		User user = createUser();
		when(userUtil.getCurrentAuthenticatedUser()).thenReturn(user);
		when(problemRepository.findById(problemId)).thenReturn(Optional.empty());

		assertThrows(
				ProblemNotFoundException.class,
				() -> submissionService.submitCode(request),
				"Problem with ID " + problemId + " not found");
		verifyNoInteractions(submissionRepository);
		verifyNoInteractions(submissionPublisher);
	}

	@Test
	void updateStatus_shouldUpdateSubmissionStatus() {
		Long submissionId = 1L;
		Submission submission = createSubmissionWithId(submissionId);
		SubmissionStatus newStatus = SubmissionStatus.RUNTIME_ERROR;
		when(submissionRepository.findById(submissionId)).thenReturn(Optional.of(submission));
		Submission newSubmission = createSubmissionWithId(submissionId);
		newSubmission.setStatus(newStatus);

		submissionService.updateStatus(submissionId, newStatus);

		verify(submissionRepository).save(newSubmission);
	}

	@Test
	void updateTimeTakenAndMemoryUsed_shouldUpdateSubmissionTimeTakenAndMemoryUsed() {
		Long submissionId = 1L;
		Submission submission = createSubmissionWithId(submissionId);
		Double timeTaken = 1.2;
		Integer memoryUsed = 24;
		when(submissionRepository.findById(submissionId)).thenReturn(Optional.of(submission));
		Submission newSubmission = createSubmissionWithId(submissionId);
		newSubmission.setExecutionTimeSeconds(timeTaken);
		newSubmission.setMemoryUsedMb(memoryUsed);

		submissionService.updateTimeTakenAndMemoryUsed(submissionId, timeTaken, memoryUsed);

		verify(submissionRepository).save(newSubmission);
	}

	private Submission createSubmissionWithId(Long submissionId) {
		Submission submission = createSubmission();
		submission.setId(submissionId);
		return submission;
	}
}
