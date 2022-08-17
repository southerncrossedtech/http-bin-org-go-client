package httpbin

import (
	"context"
)

// HttpMethodsService manages the interactions for httpbin http-methods.
type HttpMethodsService interface {
	// List returns a pager to paginate plans. PagerOptions are used to optionally
	// filter the results.
	//
	// https://dev.recurly.com/docs/list-plans
	// List(opts *PagerOptions) Pager

	// Get retrieves a plan. If the plan does not exist,
	// a nil plan and nil error are returned.
	//
	// https://dev.recurly.com/docs/lookup-plan-details
	Get(ctx context.Context) (*HttpBin, error)

	// Create a new subscription plan.
	//
	// https://dev.recurly.com/docs/create-plan
	// Create(ctx context.Context, p Plan) (*Plan, error)

	// Update the pricing or details for a plan. Existing subscriptions will
	// remain at the previous renewal amounts.
	//
	// https://dev.recurly.com/docs/update-plan
	// Update(ctx context.Context, code string, p Plan) (*Plan, error)

	// Delete makes a plan inactive. New subscriptions cannot be created
	// from inactive plans.
	//
	// https://dev.recurly.com/docs/delete-plan
	// Delete(ctx context.Context, code string) error
}

// HttpBin is the basic response
type HttpBin struct {
	Headers `json:"headers"`
	URL     string `json:"url"`
}

type Headers struct {
	XAPIVersion    string `json:"x-api-version,omitempty"`
	Authorization  string `json:"authorization,omitempty"`
	Accept         string `json:"accept,omitempty"`
	AcceptEncoding string `json:"accept-encoding,omitempty"`
	AcceptLanguage string `json:"accept-language,omitempty"`
	DNT            string `json:"dnt,omitempty"`
	Host           string `json:"host,omitempty"`
	Referer        string `json:"referer,omitempty"`
	UserAgent      string `json:"user-agent,omitempty"`
}

var _ HttpMethodsService = &httpMethodsImpl{}

// httpMethodsImpl implements HttpMethodsService.
type httpMethodsImpl serviceImpl

// func (s *plansImpl) List(opts *PagerOptions) Pager {
// 	return s.client.newPager("GET", "/plans", opts)
// }

func (s *httpMethodsImpl) Get(ctx context.Context) (*HttpBin, error) {
	req, err := s.client.newRequest(ctx, "GET", "/get", nil)
	if err != nil {
		return nil, err
	}

	var dst HttpBin
	if _, err := s.client.do(ctx, req, &dst); err != nil {
		// TODO: We'll add error handling soon
		// if e, ok := err.(*ClientError); ok && e.Response.StatusCode == http.StatusNotFound {
		// 	return nil, nil
		// }
		return nil, err
	}

	return &dst, nil
}

// func (s *plansImpl) Create(ctx context.Context, p Plan) (*Plan, error) {
// 	req, err := s.client.newRequest("POST", "/plans", p)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var dst Plan
// 	if _, err := s.client.do(ctx, req, &dst); err != nil {
// 		return nil, err
// 	}
// 	return &dst, nil
// }

// func (s *plansImpl) Update(ctx context.Context, code string, p Plan) (*Plan, error) {
// 	path := fmt.Sprintf("/plans/%s", code)
// 	req, err := s.client.newRequest("PUT", path, p)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var dst Plan
// 	if _, err := s.client.do(ctx, req, &dst); err != nil {
// 		return nil, err
// 	}
// 	return &dst, nil
// }

// func (s *plansImpl) Delete(ctx context.Context, code string) error {
// 	path := fmt.Sprintf("/plans/%s", code)
// 	req, err := s.client.newRequest("DELETE", path, nil)
// 	if err != nil {
// 		return err
// 	}

// 	_, err = s.client.do(ctx, req, nil)
// 	return err
// }
