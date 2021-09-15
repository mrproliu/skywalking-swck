// Licensed to Apache Software Foundation (ASF) under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Apache Software Foundation (ASF) licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package v1alpha1

import (
	"context"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/apache/skywalking-swck/pkg/operator/injector"
)

// log is for logging in this package.
var javaagentlog = logf.Log.WithName("javaagent")

// nolint: lll
// +kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=mpod.kb.io

// Javaagent injects java agent into Pods
type Javaagent struct {
	Client  client.Client
	decoder *admission.Decoder
}

// Handle will process every coming pod under the
// specified namespace which labeled "swck-injection=enabled"
func (r *Javaagent) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := &corev1.Pod{}

	err := r.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// set Annotations to avoid repeated judgments
	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}
	// initialize all annotation types that can be overridden
	anno, err := injector.NewAnnotations()
	if err != nil {
		javaagentlog.Error(err, "get NewAnnotations error")
	}
	// initialize Annotations to store the overlaied value
	ao := injector.NewAnnotationOverlay()
	// initialize SidecarInjectField and get injected strategy from annotations
	s := injector.NewSidecarInjectField()
	// initialize InjectProcess as a call chain
	ip := injector.NewInjectProcess(ctx, s, anno, ao, pod, req, javaagentlog, r.Client)
	// do real injection
	return ip.Run()
}

// Javaagent implements admission.DecoderInjector.
// A decoder will be automatically injected.

// InjectDecoder injects the decoder.
func (r *Javaagent) InjectDecoder(d *admission.Decoder) error {
	r.decoder = d
	return nil
}