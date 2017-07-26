/*
 * Licensed Materials - Property of tenxcloud.com
 * (C) Copyright 2017 TenxCloud. All Rights Reserved.
 */

package plugin

type WatcherPlugin interface {
	// will be called to initialize the object
	Name() string
	ForceEnabled() bool

	Init(obj interface{})
	// function when object was created
	CreateEvent(obj interface{})
	// function when object was removed
	RemoveEvent(obj interface{})
	// function when object was updated
	UpdateEvent(oldObj, newObj interface{})
}
