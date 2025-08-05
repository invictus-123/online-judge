package com.online.judge.backend.controller;

import static com.online.judge.backend.factory.ProblemFactory.createProblem;
import static com.online.judge.backend.factory.UiFactory.createProblemDetailsUi;
import static com.online.judge.backend.factory.UiFactory.createProblemSummaryUi;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.mockito.Mockito.when;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.get;
import static org.springframework.test.web.servlet.request.MockMvcRequestBuilders.post;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.content;
import static org.springframework.test.web.servlet.result.MockMvcResultMatchers.status;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.online.judge.backend.dto.filter.ProblemFilterRequest;
import com.online.judge.backend.dto.request.CreateProblemRequest;
import com.online.judge.backend.dto.request.CreateTestCaseRequest;
import com.online.judge.backend.dto.response.CreateProblemResponse;
import com.online.judge.backend.dto.response.GetProblemByIdResponse;
import com.online.judge.backend.dto.response.ListProblemsResponse;
import com.online.judge.backend.dto.ui.ProblemDetailsUi;
import com.online.judge.backend.dto.ui.ProblemSummaryUi;
import com.online.judge.backend.exception.ProblemNotFoundException;
import com.online.judge.backend.model.Problem;
import com.online.judge.backend.model.shared.ProblemDifficulty;
import com.online.judge.backend.model.shared.ProblemTag;
import com.online.judge.backend.service.ProblemService;
import java.util.List;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.test.web.servlet.MockMvc;
import org.springframework.test.web.servlet.setup.MockMvcBuilders;

@ExtendWith(MockitoExtension.class)
class ProblemControllerTest {

	private MockMvc mockMvc;

	private final ObjectMapper objectMapper = new ObjectMapper();

	@Mock
	private ProblemService problemService;

	@InjectMocks
	private ProblemController problemController;

	@BeforeEach
	void setUp() {
		mockMvc = MockMvcBuilders.standaloneSetup(problemController).build();
	}

	@Test
	void listProblems_whenRequestWithoutPageParam_returnsProblemSummaryListForFirstPage() throws Exception {
		int page = 1;
		List<Long> problemIds = List.of(1L, 2L);
		List<ProblemSummaryUi> problemSummaries =
				problemIds.stream().map(this::createProblemSummaryUiWithId).toList();
		ProblemFilterRequest expectedFilterRequest = new ProblemFilterRequest(null, null, page);
		when(problemService.listProblems(expectedFilterRequest)).thenReturn(problemSummaries);
		ListProblemsResponse expectedResponse = new ListProblemsResponse(problemSummaries);

		ResponseEntity<ListProblemsResponse> response = problemController.listProblems(page, null, null);

		assertEquals(200, response.getStatusCode().value());
		assertEquals(expectedResponse, response.getBody());

		mockMvc.perform(get("/api/v1/problems/list"))
				.andExpect(status().isOk())
				.andExpect(content().json(objectMapper.writeValueAsString(expectedResponse)));
	}

	@Test
	void listProblems_whenRequestWithPageParam_returnsProblemSummaryList() throws Exception {
		int page = 5;
		List<Long> problemIds = List.of(1L, 2L);
		List<ProblemSummaryUi> problemSummaries =
				problemIds.stream().map(this::createProblemSummaryUiWithId).toList();
		ProblemFilterRequest expectedFilterRequest = new ProblemFilterRequest(null, null, page);
		when(problemService.listProblems(expectedFilterRequest)).thenReturn(problemSummaries);
		ListProblemsResponse expectedResponse = new ListProblemsResponse(problemSummaries);

		ResponseEntity<ListProblemsResponse> response = problemController.listProblems(page, null, null);

		assertEquals(200, response.getStatusCode().value());
		assertEquals(expectedResponse, response.getBody());

		mockMvc.perform(get("/api/v1/problems/list").param("page", String.valueOf(page)))
				.andExpect(status().isOk())
				.andExpect(content().json(objectMapper.writeValueAsString(expectedResponse)));
	}

	@Test
	void getProblemById_whenProblemExists_returnsProblemDetails() throws Exception {
		Long problemId = 1L;
		ProblemDetailsUi problemDetailsUi = createProblemDetailsUiWithId(problemId);
		GetProblemByIdResponse expectedResponse = new GetProblemByIdResponse(problemDetailsUi);
		when(problemService.getProblemDetailsById(problemId)).thenReturn(problemDetailsUi);

		ResponseEntity<GetProblemByIdResponse> response = problemController.getProblemById(problemId);

		assertEquals(200, response.getStatusCode().value());
		assertEquals(expectedResponse, response.getBody());

		mockMvc.perform(get("/api/v1/problems/{id}", problemId))
				.andExpect(status().isOk())
				.andExpect(content().json(objectMapper.writeValueAsString(expectedResponse)));
	}

	@Test
	void getProblemById_whenProblemDoesNotExist_throwsNotFoundException() throws Exception {
		Long problemId = 2L;
		when(problemService.getProblemDetailsById(problemId))
				.thenThrow(new ProblemNotFoundException("Problem not found with id: " + problemId));

		mockMvc.perform(get("/api/v1/problems/{id}", problemId)).andExpect(status().isNotFound());
	}

	@Test
	void createProblem_returnsProblemDetails() throws Exception {
		CreateProblemRequest request = new CreateProblemRequest(
				"Title",
				"Statement",
				ProblemDifficulty.EASY,
				2.0,
				256,
				List.of(ProblemTag.ARRAY),
				List.of(new CreateTestCaseRequest("Sample input", "Sample output", false, "Explanation")));
		ProblemDetailsUi problemDetailsUi = createProblemDetailsUiWithId(1L);
		when(problemService.createProblem(request)).thenReturn(problemDetailsUi);
		CreateProblemResponse expectedResponse = new CreateProblemResponse(problemDetailsUi);

		ResponseEntity<CreateProblemResponse> response = problemController.createProblem(request);

		assertEquals(201, response.getStatusCode().value());
		assertEquals(expectedResponse, response.getBody());

		mockMvc.perform(post("/api/v1/problems")
						.contentType(MediaType.APPLICATION_JSON)
						.content(objectMapper.writeValueAsString(request)))
				.andExpect(status().isCreated())
				.andExpect(content().json(objectMapper.writeValueAsString(expectedResponse)));
	}

	@Test
	void listProblems_withDifficultyFilter_returnsFilteredProblems() throws Exception {
		List<ProblemSummaryUi> expectedProblems = List.of(createProblemSummaryUiWithId(1L));
		ProblemFilterRequest expectedFilterRequest = new ProblemFilterRequest(List.of(ProblemDifficulty.EASY), null, 1);
		when(problemService.listProblems(expectedFilterRequest)).thenReturn(expectedProblems);

		ListProblemsResponse expectedResponse = new ListProblemsResponse(expectedProblems);

		mockMvc.perform(get("/api/v1/problems/list").param("page", "1").param("difficulties", "EASY"))
				.andExpect(status().isOk())
				.andExpect(content().json(objectMapper.writeValueAsString(expectedResponse)));
	}

	@Test
	void listProblems_withTagFilter_returnsFilteredProblems() throws Exception {
		List<ProblemSummaryUi> expectedProblems = List.of(createProblemSummaryUiWithId(1L));
		ProblemFilterRequest expectedFilterRequest = new ProblemFilterRequest(null, List.of(ProblemTag.ARRAY), 1);
		when(problemService.listProblems(expectedFilterRequest)).thenReturn(expectedProblems);

		ListProblemsResponse expectedResponse = new ListProblemsResponse(expectedProblems);

		mockMvc.perform(get("/api/v1/problems/list").param("page", "1").param("tags", "ARRAY"))
				.andExpect(status().isOk())
				.andExpect(content().json(objectMapper.writeValueAsString(expectedResponse)));
	}

	@Test
	void listProblems_withMultipleDifficultiesAndTags_returnsFilteredProblems() throws Exception {
		List<ProblemSummaryUi> expectedProblems =
				List.of(createProblemSummaryUiWithId(1L), createProblemSummaryUiWithId(2L));
		ProblemFilterRequest expectedFilterRequest = new ProblemFilterRequest(
				List.of(ProblemDifficulty.EASY, ProblemDifficulty.MEDIUM),
				List.of(ProblemTag.ARRAY, ProblemTag.STRING),
				1);
		when(problemService.listProblems(expectedFilterRequest)).thenReturn(expectedProblems);

		ListProblemsResponse expectedResponse = new ListProblemsResponse(expectedProblems);

		mockMvc.perform(get("/api/v1/problems/list")
						.param("page", "1")
						.param("difficulties", "EASY", "MEDIUM")
						.param("tags", "ARRAY", "STRING"))
				.andExpect(status().isOk())
				.andExpect(content().json(objectMapper.writeValueAsString(expectedResponse)));
	}

	@Test
	void listProblems_withNoFiltersButCustomPage_returnsProblemsForPage() throws Exception {
		List<ProblemSummaryUi> expectedProblems = List.of(createProblemSummaryUiWithId(1L));
		ProblemFilterRequest expectedFilterRequest = new ProblemFilterRequest(null, null, 3);
		when(problemService.listProblems(expectedFilterRequest)).thenReturn(expectedProblems);

		ListProblemsResponse expectedResponse = new ListProblemsResponse(expectedProblems);

		mockMvc.perform(get("/api/v1/problems/list").param("page", "3"))
				.andExpect(status().isOk())
				.andExpect(content().json(objectMapper.writeValueAsString(expectedResponse)));
	}

	private ProblemDetailsUi createProblemDetailsUiWithId(Long id) {
		Problem problem = createProblem();
		problem.setId(id);
		return createProblemDetailsUi(problem);
	}

	private ProblemSummaryUi createProblemSummaryUiWithId(Long id) {
		Problem problem = createProblem();
		problem.setId(id);
		return createProblemSummaryUi(problem);
	}
}
