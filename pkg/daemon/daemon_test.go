package daemon

// func TestPull(t *testing.T) {
// 	ts := []struct {
// 		name       string
// 		options    *PullOptions
// 		pkgHandler *package_handler.PackageHandler
// 		err        error
// 	}{
// 		{
// 			name: "success",
// 			options: &PullOptions{
// 				URL:     "https://github.com/NethermindEth/mock-avs.git",
// 				Version: "v0.1.0",
// 				DestDir: "/tmp",
// 			},
// 			pkgHandler: package_handler.NewPackageHandler("/tmp"),
// 			err:        nil,
// 		},
// 		{
// 			name: "error",
// 			options: &PullOptions{
// 				URL:     "https://github.com/NethermindEth/mock-avs.git",
// 				Version: "v0.1.0",
// 				DestDir: "/tmp",
// 			},
// 			pkgHandler: nil,
// 			err:        errors.New("puller error"),
// 		},
// 	}
// 	for _, tc := range ts {
// 		t.Run(tc.name, func(t *testing.T) {
// 			puller := mocks.NewMockPuller(gomock.NewController(t))
// 			puller.EXPECT().
// 				Pull(tc.options.URL, tc.options.Version, tc.options.DestDir).
// 				Return(tc.pkgHandler, tc.err)

// 			d := NewWizDaemon(puller)
// 			response, err := d.Pull(tc.options)
// 			if tc.err != nil {
// 				assert.EqualError(t, err, tc.err.Error())
// 				assert.Nil(t, response)
// 			} else {
// 				assert.NoError(t, err)
// 				assert.Equal(t, &PullResponse{}, response)
// 			}
// 		})
// 	}
// }
