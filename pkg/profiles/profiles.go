package profiles

import (
	"errors"

	pb "github.com/linkerd/linkerd2-proxy-api/go/destination"
	sp "github.com/linkerd/linkerd2/controller/gen/apis/serviceprofile/v1alpha1"
	"github.com/linkerd/linkerd2/pkg/util"
)

func ToRoute(route *sp.RouteSpec) (*pb.Route, error) {
	cond, err := ToRequestMatch(route.Condition)
	if err != nil {
		return nil, err
	}
	rcs := make([]*pb.ResponseClass, 0)
	for _, rc := range route.Responses {
		pbRc, err := ToResponseClass(rc)
		if err != nil {
			return nil, err
		}
		rcs = append(rcs, pbRc)
	}
	return &pb.Route{
		Condition:       cond,
		ResponseClasses: rcs,
		MetricsLabels:   map[string]string{"route": route.Name},
	}, nil
}

func ToResponseClass(rc *sp.ResponseClass) (*pb.ResponseClass, error) {
	cond, err := ToResponseMatch(rc.Condition)
	if err != nil {
		return nil, err
	}
	return &pb.ResponseClass{
		Condition: cond,
		IsFailure: !rc.IsSuccess,
	}, nil
}

func ToResponseMatch(rspMatch *sp.ResponseMatch) (*pb.ResponseMatch, error) {
	if rspMatch == nil {
		return nil, errors.New("missing response match")
	}
	err := ValidateResponseMatch(rspMatch)
	if err != nil {
		return nil, err
	}
	if rspMatch.All != nil {
		all := make([]*pb.ResponseMatch, 0)
		for _, m := range rspMatch.All {
			pbM, err := ToResponseMatch(m)
			if err != nil {
				return nil, err
			}
			all = append(all, pbM)
		}
		return &pb.ResponseMatch{
			Match: &pb.ResponseMatch_All{
				All: &pb.ResponseMatch_Seq{
					Matches: all,
				},
			},
		}, nil
	}

	if rspMatch.Any != nil {
		any := make([]*pb.ResponseMatch, 0)
		for _, m := range rspMatch.Any {
			pbM, err := ToResponseMatch(m)
			if err != nil {
				return nil, err
			}
			any = append(any, pbM)
		}
		return &pb.ResponseMatch{
			Match: &pb.ResponseMatch_Any{
				Any: &pb.ResponseMatch_Seq{
					Matches: any,
				},
			},
		}, nil
	}

	if rspMatch.Status != nil {
		return &pb.ResponseMatch{
			Match: &pb.ResponseMatch_Status{
				Status: &pb.HttpStatusRange{
					Max: rspMatch.Status.Max,
					Min: rspMatch.Status.Min,
				},
			},
		}, nil
	}

	if rspMatch.Not != nil {
		not, err := ToResponseMatch(rspMatch.Not)
		if err != nil {
			return nil, err
		}
		return &pb.ResponseMatch{
			Match: &pb.ResponseMatch_Not{
				Not: not,
			},
		}, nil
	}

	return nil, errors.New("A response match must have a field set")
}

func ToRequestMatch(reqMatch *sp.RequestMatch) (*pb.RequestMatch, error) {
	if reqMatch == nil {
		return nil, errors.New("missing request match")
	}
	err := ValidateRequestMatch(reqMatch)
	if err != nil {
		return nil, err
	}
	if reqMatch.All != nil {
		all := make([]*pb.RequestMatch, 0)
		for _, m := range reqMatch.All {
			pbM, err := ToRequestMatch(m)
			if err != nil {
				return nil, err
			}
			all = append(all, pbM)
		}
		return &pb.RequestMatch{
			Match: &pb.RequestMatch_All{
				All: &pb.RequestMatch_Seq{
					Matches: all,
				},
			},
		}, nil
	}

	if reqMatch.Any != nil {
		any := make([]*pb.RequestMatch, 0)
		for _, m := range reqMatch.Any {
			pbM, err := ToRequestMatch(m)
			if err != nil {
				return nil, err
			}
			any = append(any, pbM)
		}
		return &pb.RequestMatch{
			Match: &pb.RequestMatch_Any{
				Any: &pb.RequestMatch_Seq{
					Matches: any,
				},
			},
		}, nil
	}

	if reqMatch.Method != "" {
		return &pb.RequestMatch{
			Match: &pb.RequestMatch_Method{
				Method: util.ParseMethod(reqMatch.Method),
			},
		}, nil
	}

	if reqMatch.Not != nil {
		not, err := ToRequestMatch(reqMatch.Not)
		if err != nil {
			return nil, err
		}
		return &pb.RequestMatch{
			Match: &pb.RequestMatch_Not{
				Not: not,
			},
		}, nil
	}

	if reqMatch.Path != "" {
		return &pb.RequestMatch{
			Match: &pb.RequestMatch_Path{
				Path: &pb.PathMatch{
					Regex: reqMatch.Path,
				},
			},
		}, nil
	}

	return nil, errors.New("A request match must have a field set")
}

func ValidateRequestMatch(reqMatch *sp.RequestMatch) error {
	tooManyKindsErr := errors.New("A request match may not have more than two fields set")
	matchKindSet := false
	if reqMatch.All != nil {
		if matchKindSet {
			return tooManyKindsErr
		}
		matchKindSet = true
		for _, child := range reqMatch.All {
			err := ValidateRequestMatch(child)
			if err != nil {
				return err
			}
		}
	}
	if reqMatch.Any != nil {
		if matchKindSet {
			return tooManyKindsErr
		}
		matchKindSet = true
		for _, child := range reqMatch.Any {
			err := ValidateRequestMatch(child)
			if err != nil {
				return err
			}
		}
	}
	if reqMatch.Method != "" {
		if matchKindSet {
			return tooManyKindsErr
		}
		matchKindSet = true
	}
	if reqMatch.Not != nil {
		if matchKindSet {
			return tooManyKindsErr
		}
		matchKindSet = true
		err := ValidateRequestMatch(reqMatch.Not)
		if err != nil {
			return err
		}
	}
	if reqMatch.Path != "" {
		if matchKindSet {
			return tooManyKindsErr
		}
		matchKindSet = true
	}

	if !matchKindSet {
		return errors.New("A request match must have a field set")
	}

	return nil
}

func ValidateResponseMatch(rspMatch *sp.ResponseMatch) error {
	tooManyKindsErr := errors.New("A response match may not have more than two fields set")
	invalidRangeErr := errors.New("Range maximum cannot be smaller than minimum")
	matchKindSet := false
	if rspMatch.All != nil {
		if matchKindSet {
			return tooManyKindsErr
		}
		matchKindSet = true
		for _, child := range rspMatch.All {
			err := ValidateResponseMatch(child)
			if err != nil {
				return err
			}
		}
	}
	if rspMatch.Any != nil {
		if matchKindSet {
			return tooManyKindsErr
		}
		matchKindSet = true
		for _, child := range rspMatch.Any {
			err := ValidateResponseMatch(child)
			if err != nil {
				return err
			}
		}
	}
	if rspMatch.Status != nil {
		if matchKindSet {
			return tooManyKindsErr
		}
		if rspMatch.Status.Max != 0 && rspMatch.Status.Min != 0 && rspMatch.Status.Max < rspMatch.Status.Min {
			return invalidRangeErr
		}
		matchKindSet = true
	}
	if rspMatch.Not != nil {
		if matchKindSet {
			return tooManyKindsErr
		}
		matchKindSet = true
		err := ValidateResponseMatch(rspMatch.Not)
		if err != nil {
			return err
		}
	}

	if !matchKindSet {
		return errors.New("A response match must have a field set")
	}

	return nil
}
