package services

import (
	"context"

	"pentagi/pkg/server/response"

	"github.com/99designs/gqlgen/graphql"
	"github.com/sirupsen/logrus"
	"github.com/vektah/gqlparser/v2/gqlerror"
)

type graphqlLanguageContextKey struct{}

func withGraphQLLanguage(ctx context.Context, acceptLanguage string) context.Context {
	return context.WithValue(ctx, graphqlLanguageContextKey{}, acceptLanguage)
}

func graphQLLanguage(ctx context.Context) string {
	acceptLanguage, _ := ctx.Value(graphqlLanguageContextKey{}).(string)
	return acceptLanguage
}

func localizedGraphQLErrorPresenter(ctx context.Context, err error) *gqlerror.Error {
	presented := graphql.DefaultErrorPresenter(ctx, err)
	localized := response.LocalizedGraphQLErrorMessage(graphQLLanguage(ctx), presented.Message)
	if localized != presented.Message {
		logrus.WithContext(ctx).WithError(err).Debug("localized graphql error response")
		presented.Message = localized
	}
	return presented
}
