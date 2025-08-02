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
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.PageImpl;
import org.springframework.data.domain.PageRequest;
import org.springframework.data.domain.Pageable;
import org.springframework.data.domain.Sort;

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

	@Test
	void listSubmissions_shouldReturnPaginatedSubmissions() {
		int page = 1;
		List<Submission> submissions = List.of(createSubmissionWithId(1L), createSubmissionWithId(2L));
		Pageable pageable =
				PageRequest.of(page - 1, PAGE_SIZE, Sort.by("submittedAt").descending());
		Page<Submission> submissionPage = new PageImpl<>(submissions, pageable, submissions.size());
		when(submissionRepository.findAll(pageable)).thenReturn(submissionPage);

		List<SubmissionSummaryUi> result = submissionService.listSubmissions(page);

		assertNotNull(result);
		assertEquals(2, result.size());
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
	}

	private Submission createSubmissionWithId(Long submissionId) {
		Submission submission = createSubmission();
		submission.setId(submissionId);
		return submission;
	}
}
